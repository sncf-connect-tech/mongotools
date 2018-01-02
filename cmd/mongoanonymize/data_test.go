package main

import "gopkg.in/mgo.v2/bson"

func CreateOrder(nbServiceItem int) bson.M {

	customers := make([]interface{}, 1)
	customer := bson.M{
		"iuc":       "myiuc!",
		"firstname": "homer",
		"lastname":  "simpson",
	}
	customers[0] = customer

	ownerCustomer := bson.M{
		"iuc": "myiuc!",
	}

	serviceItems := make([]interface{}, nbServiceItem)
	for i := 0; i < nbServiceItem; i++ {
		serviceItems[i] = CreateServiceItem()
	}

	return bson.M{
		"customers":     customers,
		"serviceItems":  serviceItems,
		"ownerCustomer": ownerCustomer,
	}
}

func CreateServiceItem() bson.M {
	contracts := make([]interface{}, 1)
	contract := bson.M{
		"holder": bson.M{
			"firstName": "chief",
			"lastName":  "wiggum",
		},
	}
	contracts[0] = contract

	contactInformation := bson.M{
		"firstname":            "bart",
		"name":                 "simpson",
		"address1":             "springfield",
		"mobilePhoneNumber":    "010101010101",
		"emailAddress":         "bart@simpson.fr",
		"landlinePhoneNumbers": "yo",
	}

	vehicleOwner := bson.M{
		"iuc":       "myiuc!",
		"firstName": "Barney",
		"lastName":  "Gamble",
	}

	clientInformations := bson.M{
		"fidNumber": "1234567890",
	}

	clientCard := bson.M{
		"cardNumber": "1234567890",
	}

	sncfAgentAdditionalProperties := bson.M{
		"agentId": "DFERT23",
	}

	passengers := make([]interface{}, 2)
	passenger1 := bson.M{
		"iuc":                "myIUC?",
		"firstName":          "bart",
		"lastName":           "Simpson",
		"mobilePhoneNumber":  "06141725282435",
		"emailAddress":       "bart.simpson.fr",
		"clientInformations": clientInformations,
	}
	passenger2 := bson.M{
		"firstName":                     "lisa",
		"lastName":                      "Simpson",
		"mobilePhoneNumber":             "06141725282499",
		"emailAddress":                  "lisa.simpson.fr",
		"fceNumber":                     "1234567890",
		"clientCard":                    clientCard,
		"sncfAgentAdditionalProperties": sncfAgentAdditionalProperties,
		"birthDate":                     "1993-05-25T00:00:00+02:00",
	}
	passengers[0] = passenger1
	passengers[1] = passenger2

	vehicles := make([]interface{}, 1)
	vehicle := bson.M{
		"plateNumber": "AV-345-RV",
	}
	vehicles[0] = vehicle

	return bson.M{
		"contactInformation":          contactInformation,
		"railTransportationContracts": contracts,
		"vehicleOwner":                vehicleOwner,
		"passengers":                  passengers,
		"vehicles":                    vehicles,
	}
}
