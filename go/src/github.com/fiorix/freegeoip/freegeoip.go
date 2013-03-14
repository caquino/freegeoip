package main

import (
    "fmt"
    "regexp"
    "encoding/json"
    "encoding/xml"
    "net/http"
)

var apiValidator = regexp.MustCompile("/(json|csv|xml)/(.*)")

func apiHandler(ipdb *IpDb, fn func (http.ResponseWriter, *http.Request, IpMessage)) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {

        w.Header().Set("Access-Control-Allow-Origin", "*")

        vars := apiValidator.FindStringSubmatch(r.URL.Path)
        if len(vars) < 1 {
            http.NotFound(w, r)
            return
        }

        reqc := make (chan IpMessage)
        ipdb.Query(vars[2], reqc)
        loc := <-reqc

        if !loc.Found {
            http.NotFound(w, r)
            return
        }

        fn(w, r, loc)
    }
}

func jsonHandler(w http.ResponseWriter, r *http.Request, ipresp IpMessage) {
    w.Header().Set("Content-Type", "application/json")
    resp, err := json.Marshal(ipresp)
    if err != nil {
        http.NotFound(w, r)
        return
    }
    fmt.Fprintf(w, "%s\n", resp)
}


func xmlHandler(w http.ResponseWriter, r *http.Request, ipresp IpMessage) {
    w.Header().Set("Content-Type", "application/xml")
    resp, err := xml.MarshalIndent(ipresp, " ", "  ")
    if err != nil {
        http.NotFound(w, r)
        return
    }
    fmt.Fprintf(w, "%s\n", resp)
}


func csvHandler(w http.ResponseWriter, r *http.Request, ipresp IpMessage) {
    w.Header().Set("Content-Type", "application/csv")
    fmt.Fprintf(w, "\"%s\",\"%s\",\"%s\",\"%s\",\"%s\",\"%s\",\"%s\",\"%f\",\"%f\",\"%s\",\"%s\"\n",
                    ipresp.Ip,
                    ipresp.CountryCode,
                    ipresp.CountryName,
                    ipresp.RegionCode,
                    ipresp.RegionName,
                    ipresp.CityName,
                    ipresp.ZipCode,
                    ipresp.Latitude,
                    ipresp.Longitude,
                    ipresp.MetroCode,
                    ipresp.AreaCode)
}

func main() {
    var ipdb, err = OpenIpDb("./ipdb.db")
    if err != nil {
        fmt.Println("Error opening database")
    }
    http.HandleFunc("/json/", apiHandler(ipdb, jsonHandler));
    http.HandleFunc("/xml/", apiHandler(ipdb, xmlHandler));
    http.HandleFunc("/csv/", apiHandler(ipdb, csvHandler));
    fmt.Println("FreeGeoIP starting")
    http.ListenAndServe(":8080", nil)
}

