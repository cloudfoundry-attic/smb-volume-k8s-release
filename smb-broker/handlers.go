package main

import "net/http"

func BrokerHandler() http.Handler {
	return http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`{
  "services": [{
    "bindable": true,
    "description": "SMB Broker for k8s",
    "id": "0A789746-596F-4CEA-BFAC-A0795DA056E3",
    "name": "smb-broker",
    "plan_updateable": false,
    "plans": [{
      "description": "The only SMB plan",
      "id": "plan-id",
      "name": "Existing",
      "metadata": {
        "displayName": "SMB"
      },
      "schemas": {
        "service_instance": {
          "create": {
            "parameters": {
              "$schema": "http://json-schema.org/draft-04/schema#",
              "type": "object",
              "properties": {}
            }
          },
          "update": {
            "parameters": {
              "$schema": "http://json-schema.org/draft-04/schema#",
              "type": "object",
              "properties": {}
            }
          }
        },
        "service_binding": {
          "create": {
            "parameters": {
              "$schema": "http://json-schema.org/draft-04/schema#",
              "type": "object",
              "properties": {}
            }
          }
        }
      }
    }],
    "metadata": {
      "displayName": "SMB",
      "longDescription": "Long description",
      "documentationUrl": "http://thedocs.com",
      "supportUrl": "http://helpme.no"
    },
    "tags": [
      "pivotal",
      "smb",
      "volume-services"
    ]
  }]
}`))
		if err != nil {
			panic(err)
		}
	}))
}