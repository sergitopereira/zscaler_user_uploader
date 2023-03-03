package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"
	"github.com/zepryspet/gozscaler/zia"
)

type Zid interface {
	GetID() (string, int)
}

// GetIDs is a generic function that receives an arrray object and return a map with the name as key and ID as value
func GetIDs[K Zid](obj []K) map[string]int {
	//Creating map
	m := make(map[string]int)
	//Iterating
	for _, v := range obj {
		name, id := v.GetID()
		m[name] = id
	}
	return m
}

// Read csv and return

func ReadCsv() [][]string {
	f, err := os.Open(os.Getenv("ZS_USER_DB"))
	if err != nil {
		log.Fatal(err)
	}

	// Close the file at the end of the program
	defer f.Close()

	// read csv values using csv.Reader
	csvReader := csv.NewReader(f)
	// Remove the first line
	csvReader.Read()
	rows, err := csvReader.ReadAll()
	if err != nil {
		log.Println("Cannot read CSV file:", err)
	}
	return rows

}

func main() {
	datacsv := ReadCsv()
	z, e := zia.NewClient(os.Getenv("ZS_CLOUD"), os.Getenv("ZS_ADMIN"), os.Getenv("ZS_PASSWORD"), os.Getenv("ZS_APIKEY"))
	if e != nil {
		fmt.Println(e.Error())

	}
	// obtain configured groups
	log.Println("->Obtaining User Groups")
	obj, e := z.GetGroups()
	if e != nil {
		fmt.Println(e.Error())
	}
	cfg_groups := GetIDs(obj)
	//fmt.Println(cfg_groups)

	// obtain configured Departments
	log.Println("->Obtaining User Departments")
	obj2, e := z.GetDeparments()
	if e != nil {
		fmt.Println(e.Error())
	}
	cfg_departments := GetIDs(obj2)
	//fmt.Println(cfg_departments)

	for _, row := range datacsv {

		tmp := zia.User{}
		grp := zia.UserGroup{}
		dep := zia.Department{}
		tmp_groups := []zia.UserGroup{}

		// check group exists, otherwise use default Service Admin
		if  len(row[8]) >1 {
			////  split and ger group ids
			for _, g := range strings.Split(row[8], "$") {
				grp.ID = cfg_groups[g]
				tmp_groups = append(tmp_groups, grp)
			}

		} else {
			grp.ID = cfg_groups["Service Admin"]
			tmp_groups = append(tmp_groups, grp)
		}

		// check Department exists, otherwise use Service Admin
		if _, ok := cfg_departments[row[9]]; ok {
			dep.ID = cfg_departments[row[9]]
		} else {
			dep.ID = cfg_departments["Service Admin"]
		}
		tmp.Name = row[4]
		tmp.Email = row[6]
		tmp.Password = row[3]
		tmp.AdminUser = false
		tmp.Department = dep
		tmp.Groups = tmp_groups
		resp, e := z.AddUser(tmp)
		if e != nil {
			fmt.Println(e.Error())
		}
		log.Println("->User ", tmp.Email, "created with id", resp)

	}
	log.Println("Activating chages")
	z.Activate()

}
