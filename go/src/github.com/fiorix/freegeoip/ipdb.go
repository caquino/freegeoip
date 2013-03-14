package main

import (
    "encoding/xml"
    "fmt"
    "net"
    "bytes"
    "encoding/binary"
    "database/sql"
    _   "github.com/mattn/go-sqlite3"
)

type IpDb struct {
    db *sql.DB
}

type IpMessage struct {
    XMLName   xml.Name `json:"-" xml:"Response"`
    Found bool `json:"-" xml:"-"`
    CityName string `json:"city" xml:"City"`
    RegionCode string `json:"region_code"`
    RegionName string `json:"region_name"`
    AreaCode string `json:"areacode"`
    Ip string `json:"ip"`
    ZipCode string `json:"zipcode"`
    Longitude float32 `json:"longitude"`
    Latitude float32 `json:"latitude"`
    CountryName string `json:"contry_name"`
    CountryCode string `json:"country_code"`
    MetroCode string `json:"metro_code"`
}

func OpenIpDb(dbfile string) (*IpDb, error) {
    var db IpDb
    var err error
    db.db , err = sql.Open("sqlite3", dbfile)
    if err != nil {
        return nil, err
    }
    return &db, nil
}


func (db *IpDb) String() string {
    return "yep"
}

func (db *IpDb) Query(ip string, c chan IpMessage) {
    go func() {
        var ret IpMessage
        lIP := net.ParseIP(ip)
        ret.Ip = ip
        query :=     "SELECT " +
        "  city_location.country_code, country_blocks.country_name, " +
        "  city_location.region_code, region_names.region_name, " +
        "  city_location.city_name, city_location.postal_code, " +
        "  city_location.latitude, city_location.longitude, " +
        "  city_location.metro_code, city_location.area_code " +
        "FROM city_blocks " +
        "  NATURAL JOIN city_location " +
        "  INNER JOIN country_blocks ON " +
        "    city_location.country_code = country_blocks.country_code " +
        "  INNER JOIN region_names ON " +
        "    city_location.country_code = region_names.country_code " +
        "    AND " +
        "    city_location.region_code = region_names.region_code " +
        "WHERE city_blocks.ip_start <= ? " +
        "ORDER BY city_blocks.ip_start DESC LIMIT 1"
        stmt, err := db.db.Prepare(query)

        if err != nil {
            fmt.Println("error preparing ", err)
            c <- ret
            return
        }

        defer stmt.Close()


        b := bytes.NewBuffer(lIP.To4())
        var queryIP uint32
        binary.Read(b, binary.BigEndian, &queryIP)

        err = stmt.QueryRow(queryIP).Scan(
            &ret.CountryCode, &ret.CountryName,
            &ret.RegionCode, &ret.RegionName,
            &ret.CityName, &ret.ZipCode,
            &ret.Latitude, &ret.Longitude,
            &ret.MetroCode, &ret.AreaCode)
        if err == nil {
            ret.Found = true
        }
        c <- ret
    }()
}
