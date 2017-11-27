package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"

	"code.cloudfoundry.org/cli/plugin"
)

// simpleClient is a simple CloudFoundry client
type simpleClient struct {
	// API url, ie "https://api.system.example.com"
	API string

	// Authorization header, ie "bearer eyXXXXX"
	Authorization string
}

// Get makes a GET request, where r is the relative path, and rv is json.Unmarshalled to
func (sc *simpleClient) Get(r string, rv interface{}) error {
	log.Printf("GET %s%s\n", sc.API, r)
	req, err := http.NewRequest(http.MethodGet, sc.API+r, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", sc.Authorization)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("bad status code")
	}

	return json.NewDecoder(resp.Body).Decode(rv)
}

type upsUtil struct{}

func printCups(name string, kv map[string]string) (string, error) {
	jv, err := json.MarshalIndent(kv, "", "    ")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("cf create-user-provided-service %s -p @<(cat <<EOF\n%s\nEOF\n)", name, jv), nil
}

func (c *upsUtil) Run(cliConnection plugin.CliConnection, args []string) {
	if args[0] == "ups-util" {
		apps := make(map[string]string)
		if len(args) == 1 {
			// all apps
			rawApps, err := cliConnection.GetApps()
			if err != nil {
				log.Fatal(err)
			}
			for _, app := range rawApps {
				apps[app.Guid] = app.Name
			}
		} else {
			for _, name := range args[1:] {
				app, err := cliConnection.GetApp(name)
				if err != nil {
					log.Fatal(err)
				}
				apps[app.Guid] = app.Name
			}
		}

		at, err := cliConnection.AccessToken()
		if err != nil {
			log.Fatal(err)
		}

		api, err := cliConnection.ApiEndpoint()
		if err != nil {
			log.Fatal(err)
		}

		client := &simpleClient{
			API:           api,
			Authorization: at,
		}

		appToKeyToVal := make(map[string]map[string]string)
		for appGUID, appName := range apps {
			var appStuff struct {
				EnvVars map[string]string `json:"environment_json"`
			}
			err = client.Get(fmt.Sprintf("/v2/apps/%s/env", appGUID), &appStuff)
			if err != nil {
				log.Fatal(err)
			}
			appToKeyToVal[appName] = appStuff.EnvVars
		}

		keysToVals := make(map[string]map[string]bool)
		keysToApps := make(map[string]map[string]bool)
		for app, kv := range appToKeyToVal {
			for k, v := range kv {
				if keysToVals[k] == nil {
					keysToVals[k] = make(map[string]bool)
				}
				keysToVals[k][v] = true

				if keysToApps[k] == nil {
					keysToApps[k] = make(map[string]bool)
				}
				keysToApps[k][app] = true
			}
		}

		commonEnv := make(map[string]string)
		for k, vals := range keysToVals {
			if len(vals) == 1 && len(keysToApps[k]) > 1 {
				for val := range vals {
					commonEnv[k] = val
				}
			}
		}

		if len(commonEnv) != 0 {
			ups, err := printCups("ups-shared", commonEnv)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(ups)
		}

		var appNames []string
		for app := range appToKeyToVal {
			appNames = append(appNames, app)
		}
		sort.Strings(appNames)

		for _, app := range appNames {
			appOnly := make(map[string]string)
			for k, v := range appToKeyToVal[app] {
				if commonEnv[k] != v {
					appOnly[k] = v
				}
			}

			if len(appOnly) != 0 {
				ups, err := printCups("ups-"+app, appOnly)
				if err != nil {
					log.Fatal(err)
				}
				fmt.Println(ups)
			}
		}
	}
}

func (c *upsUtil) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "UPS Util",
		Version: plugin.VersionType{
			Major: 0,
			Minor: 1,
			Build: 0,
		},
		MinCliVersion: plugin.VersionType{
			Major: 6,
			Minor: 7,
			Build: 0,
		},
		Commands: []plugin.Command{
			{
				Name:     "ups-util",
				HelpText: "Converts env vars to create-user-provided servies (just prints all in current space)",
				UsageDetails: plugin.Usage{
					Usage: "ups-util\n   cf ups-util",
				},
			},
		},
	}
}

func main() {
	plugin.Start(&upsUtil{})
}
