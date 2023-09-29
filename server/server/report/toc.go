package report

import (
	"slices"
	"strings"

	"github.com/elmasy-com/columbus/db"
	"github.com/elmasy-com/columbus/frontend"
)

func buildSubList(doms []db.Domain) *frontend.SubData {

	root := new(frontend.SubData)
	root.Childs = make([]*frontend.SubData, 0)

	for i := range doms {

		if doms[i].Sub == "" {
			root.Domain = doms[i].FullDomain()
			continue
		}

		current := root

		parts := strings.Split(doms[i].Sub, ".")
		slices.Reverse(parts)

		for j := range parts {

			// Add the full domain to the latest part, to use as an id in the report
			d := ""
			if j == len(parts)-1 {
				d = doms[i].String()
			}

			current = current.Add(parts[j], d)
		}
	}

	return root
}
