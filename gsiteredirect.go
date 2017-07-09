// Copyright 2016 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// This package redirects requests to appropriate Google Site for your domain
// The main goal of this package to help avoid visitors from having to type
// https://sites.google.com/a/your.domain.com/GOOGLESITENAME or
// https://sites.google.com/your.domain.com/GOOGLESITENAME to go to a
// Google Site created within Google Apps for your domain

package gsiteredirect

import (
	"html/template"
	"net/http"
	"strings"
	"regexp"
	"time"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
	"google.golang.org/appengine/memcache"
)

var siteTmpl = template.Must(template.ParseFiles("site.html"))

func init() {
	http.HandleFunc("/", handleRedirect)
}

func handleRedirect(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	client := urlfetch.Client(c)

	requestUri := r.RequestURI
	matched, err := regexp.MatchString("/.+", requestUri)
	if matched == false {
		if err := siteTmpl.Execute(w, nil); err != nil {
			log.Errorf(c, "Error executing siteTmpl template: %s", err)
			http.Error(w, "Error executing template", http.StatusInternalServerError)
		}
		return
	} else if err != nil {
		log.Errorf(c, "Error performing regex match on URI: %s", err)
		http.Error(w, "Error performing regex match", http.StatusInternalServerError)
		return
	}
// Split the request URI into upto four parts delimeted by /
	splitUri := strings.SplitN(requestUri, "/", 4)
	gSiteName := splitUri[1]
// Let's check to see if the URI already exists in our cache.
// If it does, then we will redirect the user instead of making a HTTP request
	item, err := memcache.Get(c, gSiteName)
	if err != nil && err != memcache.ErrCacheMiss {
		log.Errorf(c, "Error querying memcache: %s", err)
	} else if err != nil && err == memcache.ErrCacheMiss {
		log.Infof(c, "memcache miss for key: %s", gSiteName)
	} else {
		cacheGSite := string(item.Value)
		log.Infof(c, "Redirecting request URI %s to cached destination URI %s", requestUri, cacheGSite)
		http.Redirect(w, r, cacheGSite, 302)
		return
	}

// Looks like we weren't able to find the URI in our cache
// We will now construct the full URI for the request
// And then make HTTP requests to Google to see if the desired Google Site is in
// Old or new Google Sites	
	oldGSite := oldGSiteBase + gSiteName
	newGSite := newGSiteBase + gSiteName

// We are going to check old Google Site first
	responseOld, errOld := client.Get(oldGSite)
	if errOld == nil && responseOld.StatusCode == 200  {
		log.Infof(c, "Received response: %s", responseOld.Status)
// Since we received a successful response from Google Sites,
// We will store the information in memcache for 1 HOUR so the nex time someone
// Wants to visit the same site, we don't have to do a http request
		item := &memcache.Item {
			Key: gSiteName,
			Value: []byte(oldGSite),
			Expiration: 1 * time.Hour,
		}
		if err := memcache.Set(c, item); err != nil {
			log.Errorf(c, "Error saving Key: %q due to error: %s", item.Key, err)
		}
		log.Infof(c, "Redirecting request URI %s to destination URI %s", requestUri, oldGSite)
		http.Redirect(w, r, oldGSite, 302)
		return
	}
	_ = responseOld

// We are in this section because the site does not exist under old Google Sites
// We will now check new Google Sites and see if it exists there
	responseNew, errNew := client.Get(newGSite)
	if errNew == nil && responseNew.StatusCode == 200 {
		log.Errorf(c, "Received response: %s", responseNew.Status)
// Since we received a successful response from Google Sites,
// We will store the information in memcache for 1 HOUR so the nex time someone
// Wants to visit the same site, we don't have to do a http request
		item := &memcache.Item {
			Key: gSiteName,
			Value: []byte(newGSite),
			Expiration: 1 * time.Hour,
		}
		if err := memcache.Set(c, item); err != nil {
			log.Errorf(c, "Error saving Key: %q due to error: %s", item.Key, err)
		}
		log.Infof(c, "Redirecting request URI %s to destination URI %s", requestUri, newGSite)
		http.Redirect(w, r, newGSite, 302)
		return
	}
	_ = responseNew
	
//	Looks like we weren't able to find the site, so we throw a 404 and make someone sad
	log.Infof(c, "Unable to find URL at new or old Google sites: %s", oldGSite + " " + newGSite)
	http.Error(w, "404: Unable to locate Google site", http.StatusNotFound)
	return
}
