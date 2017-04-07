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

// Package quotebot demonstrates how to create an App Engine application as a
// Slack slash command.
package gsiteredirect

import (
	"html/template"
	"net/http"
	"strings"
	"regexp"
	_ "time"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

var indexTmpl = template.Must(template.ParseFiles("index.html"))
var siteTmpl = template.Must(template.ParseFiles("site.html"))

func init() {
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/s/", handleRedirect)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if err := indexTmpl.Execute(w, nil); err != nil {
		c := appengine.NewContext(r)
		log.Errorf(c, "Error executing indexTmpl template: %s", err)
	}
}

func handleRedirect(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	client := urlfetch.Client(c)
	
	requestUri := r.RequestURI
	matched, err := regexp.MatchString("/s/.+", requestUri)
	if matched == false {
		if err := siteTmpl.Execute(w, nil); err != nil {
			log.Errorf(c, "Error executing siteTmpl template: %s", err)
		}
		return
	} else if err != nil {
		log.Errorf(c, "Error performing regex match on URI: %s", err)
		http.Error(w, "Error performing regex match", http.StatusInternalServerError)
		return
	}
			
	splitUri := strings.SplitN(requestUri, "/", 4)
	
	oldGSite := "https://sites.google.com/a/umich.edu/" + splitUri[2]
	newGSite := "https://sites.google.com/umich.edu/" + splitUri[2]
	
	responseOld, errOld := client.Get(oldGSite)
	if errOld == nil && responseOld.StatusCode == 200  {
		log.Errorf(c, "Received response: %s", responseOld.Status)
		http.Redirect(w, r, oldGSite, 302)
		return
	}
	_ = responseOld

	responseNew, errNew := client.Get(newGSite)
	if errNew == nil && responseNew.StatusCode == 200 {
		log.Errorf(c, "Received response: %s", responseNew.Status)
		http.Redirect(w, r, newGSite, 302)
		return
	}
	_ = responseNew
	
//	Looks like we weren't able to find the site, so we throw a 404 and make someone sad
	log.Errorf(c, "Unable to find URL at new or old Google sites: %s", oldGSite + " " + newGSite)
	http.Error(w, "404: Unable to locate Google site", http.StatusNotFound)
	return
	
}
