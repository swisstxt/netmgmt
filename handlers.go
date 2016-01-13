package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/unrolled/render"
)

func GetNodeInfo(res http.ResponseWriter, req *http.Request) {
	r := render.New()
	vars := mux.Vars(req)

	node := vars["node"]
	if node == "" {
		r.JSON(res, http.StatusInternalServerError, "No node name provided")
		return
	}

	sets, err := Resolv(node)
	if err != nil {
		r.JSON(res, http.StatusNotFound, "Node could not be resolved")
		return
	}

	for _, s := range sets {
		for _, n := range networks {
			if n.Contains(s.Addr) {
				r.JSON(res, http.StatusOK, n)
				return
			}
		}
	}
	r.JSON(res, http.StatusNotFound, "No matching network found")
}

func GetNetworks(res http.ResponseWriter, req *http.Request) {
	r := render.New()
	r.JSON(res, http.StatusOK, networks)
}

func GetNetwork(res http.ResponseWriter, req *http.Request) {
	r := render.New()
	vars := mux.Vars(req)

	network_name := vars["net"]
	if network_name == "" {
		r.JSON(res, http.StatusInternalServerError, "No network name provided")
		return
	}

	for _, network := range networks {
		if network.Name == network_name {
			r.JSON(res, http.StatusOK, network)
			return
		}
	}
	r.JSON(res, http.StatusNotFound, "No matching network found")
}

func GetNetworkIps(res http.ResponseWriter, req *http.Request) {
	r := render.New()
	vars := mux.Vars(req)

	network_name := vars["net"]
	if network_name == "" {
		r.JSON(res, http.StatusInternalServerError, "No network name provided")
		return
	}

	for _, network := range networks {
		if network.Name == network_name {
			ips, err := network.ExpandDetailed()
			if err != nil {
				r.JSON(res, http.StatusInternalServerError, "Network could not be expanded")
				return
			}

			c := NewCheck(ips)
			c.Run()
			network.Utilization = c.utilization

			var out []*ResultSet

			for _, elem := range c.results {
				out = append(out, elem)
			}

			r.JSON(res, http.StatusOK, out)
			return

		}
	}
	r.JSON(res, http.StatusNotFound, "No matching network found")
}

func PostReservation(res http.ResponseWriter, req *http.Request) {
	r := render.New()
	vars := mux.Vars(req)

	network_name := vars["net"]
	if network_name == "" {
		r.JSON(res, http.StatusInternalServerError, "No network name provided")
		return
	}

	decoder := json.NewDecoder(req.Body)
	var l Lock
	err := decoder.Decode(&l)
	if err != nil {
		r.JSON(res, http.StatusInternalServerError, "Could not extract request body")
		return
	}

	comment := l.Comment
	if comment == "" {
		r.JSON(res, http.StatusInternalServerError, "No comment provided")
		return
	}

	for _, network := range networks {
		if network.Name == network_name {
			ips, err := network.ExpandManaged()
			if err != nil {
				r.JSON(res, http.StatusInternalServerError, "Network could net be expanded")
				return
			}

			c := NewCheck(ips)
			c.Run()

			for ip, status := range c.results {
				if status.Used() {
					delete(c.results, ip)
				}
			}

			random := rand.New(rand.NewSource(time.Now().UnixNano()))
			elem := random.Intn(len(c.results))

			i := 0
			for ip := range c.results {
				if elem == i {
					locker.Add(ip, comment, "")
					r.JSON(res, http.StatusOK, ip)
					return
				}
				i++
			}
		}
	}
	r.JSON(res, http.StatusNotFound, "No matching network found")
}

func GetConfig(res http.ResponseWriter, req *http.Request) {
	r := render.New()
	r.JSON(res, http.StatusOK, config)
}

func GetUI(res http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{
		Directory: "assets",
		Asset: func(name string) ([]byte, error) {
			return Asset(name)
		},
		AssetNames: func() []string {
			return []string{"assets/index.tmpl"}
		},
		Delims: render.Delims{"[[[", "]]]"},
	})
	r.HTML(res, http.StatusOK, "index", config)
}
