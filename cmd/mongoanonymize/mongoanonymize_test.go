package main

import (
	"github.com/voyages-sncf-technologies/mongotools/doccleaner"
	"gopkg.in/mgo.v2/bson"
	"strings"
	"testing"
	"time"
)

var (
	fakeDate, _ = time.Parse("Jan 2 15:04:05 -0700 MST 2006", "Jan 1 00:00:00 -0100 MST 1900")
)

func testFieldValueInArray(document bson.M, t *testing.T, subDoc string, fieldName string, index int, expected interface{}) {
	fieldValue := document[subDoc].([]interface{})[index].(bson.M)[fieldName]
	if fieldValue != expected {
		logError(expected, fieldValue, t, subDoc, fieldName)
	}
}

func testFieldValue(document bson.M, t *testing.T, subDoc string, fieldName string, expected string) {
	fieldValue := document[subDoc].(bson.M)[fieldName]
	if fieldValue != expected {
		logError(expected, fieldValue, t, subDoc, fieldName)
	}
}

func testFieldValueAsNil(document bson.M, t *testing.T, subDoc string, fieldName string) {
	fieldValue := document[subDoc].(bson.M)[fieldName]
	if fieldValue != nil {
		logErrorForNil(fieldValue, t, subDoc, fieldName)
	}
}

func testFieldValueInArrayAsNil(document bson.M, t *testing.T, subDoc string, fieldName string, index int) {
	fieldValue := document[subDoc].([]interface{})[index].(bson.M)[fieldName]
	if fieldValue != nil {
		logErrorForNil(fieldValue, t, subDoc, fieldName)
	}
}

func testCustomer(document bson.M, t *testing.T) {
	testFieldValueInArray(document, t, "customers", "iuc", 0, "xxx")
	testFieldValueInArray(document, t, "customers", "firstname", 0, "prenom")
	testFieldValueInArray(document, t, "customers", "lastname", 0, "nom")
}

func testOwnerCustomer(document bson.M, t *testing.T) {
	testFieldValue(document, t, "ownerCustomer", "iuc", "xxx")
}

func testContactInformation(document bson.M, t *testing.T) {
	testFieldValue(document, t, "contactInformation", "firstname", "firstname")
	testFieldValue(document, t, "contactInformation", "name", "name")
	testFieldValue(document, t, "contactInformation", "address1", "address1")
	testFieldValueAsNil(document, t, "contactInformation", "address2")
	testFieldValueAsNil(document, t, "contactInformation", "address3")
	testFieldValueAsNil(document, t, "contactInformation", "address4")
	testFieldValueAsNil(document, t, "contactInformation", "city")
	testFieldValueAsNil(document, t, "contactInformation", "zipCode")
	testFieldValueAsNil(document, t, "contactInformation", "country")
	testFieldValue(document, t, "contactInformation", "mobilePhoneNumber", "0123456789")
	testFieldValue(document, t, "contactInformation", "emailAddress", "toto@toto.fr")
	testFieldValueAsNil(document, t, "contactInformation", "landlinePhoneNumbers")
}

func testPassengers(document bson.M, t *testing.T) {

	// Pax 1
	testFieldValueInArray(document, t, "passengers", "iuc", 0, "iuc")
	testFieldValueInArray(document, t, "passengers", "firstName", 0, "firstName")
	testFieldValueInArray(document, t, "passengers", "lastName", 0, "lastName")
	testFieldValueInArray(document, t, "passengers", "mobilePhoneNumber", 0, "0123456789")
	testFieldValueInArray(document, t, "passengers", "emailAddress", 0, "toto@toto.fr")
	testFieldValueInArrayAsNil(document, t, "passengers", "fceNumber", 0)
	testFieldValueInArrayAsNil(document, t, "passengers", "mrcNumber", 0)
	testFieldValueInArrayAsNil(document, t, "passengers", "birthDate", 0)

	pax1 := document["passengers"].([]interface{})[0].(bson.M)
	testFieldValue(pax1, t, "clientInformations", "fidNumber", "xxx")

	// Pax 2
	testFieldValueInArrayAsNil(document, t, "passengers", "iuc", 1)
	testFieldValueInArray(document, t, "passengers", "firstName", 1, "firstName")
	testFieldValueInArray(document, t, "passengers", "lastName", 1, "lastName")
	testFieldValueInArray(document, t, "passengers", "mobilePhoneNumber", 1, "0123456789")
	testFieldValueInArray(document, t, "passengers", "emailAddress", 1, "toto@toto.fr")
	testFieldValueInArray(document, t, "passengers", "fceNumber", 1, "fceNumber")
	testFieldValueInArrayAsNil(document, t, "passengers", "mrcNumber", 1)
	// special case for date
	if birthDate := document["passengers"].([]interface{})[1].(bson.M)["birthDate"].(time.Time); birthDate.Unix() != fakeDate.Unix() {
		logError(fakeDate, birthDate, t, "passengers", "birthDate")
	}

	pax2 := document["passengers"].([]interface{})[1].(bson.M)
	testFieldValue(pax2, t, "clientCard", "cardNumber", "xxx")
	testFieldValue(pax2, t, "sncfAgentAdditionalProperties", "agentId", "xxx")

}

func testRailTransportationContracts(document bson.M, t *testing.T) {
	testFieldValue(document, t, "holder", "firstName", "prenom")
	testFieldValue(document, t, "holder", "lastName", "nom")
}

func testVehicles(document bson.M, t *testing.T) {
	testFieldValueInArray(document, t, "vehicles", "plateNumber", 0, "xxx")
}

func testVehicleOwner(document bson.M, t *testing.T) {
	testFieldValue(document, t, "vehicleOwner", "iuc", "xxx")
	testFieldValue(document, t, "vehicleOwner", "firstName", "prenom")
	testFieldValue(document, t, "vehicleOwner", "lastName", "nom")
}

func testServiceItems(document bson.M, t *testing.T, nbServiceItems int) {
	for i := 0; i < nbServiceItems; i++ {
		serviceItem := document["serviceItems"].([]interface{})[i].(bson.M)
		testServiceItem(serviceItem, t)
	}
}

func testServiceItem(serviceItem bson.M, t *testing.T) {
	testContactInformation(serviceItem, t)
	testPassengers(serviceItem, t)
	testRailTransportationContracts(serviceItem["railTransportationContracts"].([]interface{})[0].(bson.M), t)
	testVehicleOwner(serviceItem, t)
	testVehicles(serviceItem, t)
}

func logError(expected interface{}, result interface{}, t *testing.T, doc string, field string) {
	t.Errorf("expecting %v (%T) in document %v for field %s, got %v (%T)", expected, expected, doc, field, result, result)
}

func logErrorForNil(result interface{}, t *testing.T, doc string, field string) {
	t.Errorf("expecting nil in document % for field %s, got %s", result, doc, field)
}

func BenchmarkAnonymize(b *testing.B) {
	// given
	config := `
["customers.iuc"]
"method"="set"
"args" = ["xxx"]
["customers.firstname"]
"method"="set"
"args" = [ "prenom" ]
["customers.lastname"]
"method"="set"
"args" = [ "nom" ]
["ownerCustomer.iuc"]
"method"="set"
"args" = [ "xxx" ]

["serviceItems.contactInformation.firstname"]
"method"="set"
"args" = [ "firstname" ]
["serviceItems.contactInformation.name"]
"method"="set"
"args" = [ "name" ]
["serviceItems.contactInformation.emailAddress"]
"method"="set"
"args" = [ "toto@toto.fr" ]
["serviceItems.contactInformation.address1"]
"method"="set"
"args" = [ "address1" ]
["serviceItems.contactInformation.address2"]
"method"="set"
"args" = [ "address2" ]
["serviceItems.contactInformation.address3"]
"method"="set"
"args" = [ "address3" ]
["serviceItems.contactInformation.address4"]
"method"="set"
"args" = [ "address4" ]
["serviceItems.contactInformation.city"]
"method"="set"
"args" = [ "city" ]
["serviceItems.contactInformation.zipCode"]
"method"="set"
"args" = [ "zipCode" ]
["serviceItems.contactInformation.country"]
"method"="set"
"args" = [ "country" ]
["serviceItems.contactInformation.mobilePhoneNumber"]
"method"="set"
"args" = [ "0123456789" ]
["serviceItems.contactInformation.landlinePhoneNumbers"]
"method"="nil"
["serviceItems.passengers.emailAddress"]
"method"="set"
"args" = [ "toto@toto.fr" ]
["serviceItems.passengers.iuc"]
"method"="set"
"args" = [ "iuc" ]
["serviceItems.passengers.firstName"]
"method"="set"
"args" = [ "firstName" ]
["serviceItems.passengers.lastName"]
"method"="set"
"args" = [ "lastName" ]
["serviceItems.passengers.mobilePhoneNumber"]
"method"="set"
"args" = ["0123456789" ]
["serviceItems.passengers.fceNumber"]
"method"="set"
"args"=["fceNumber"]
["serviceItems.passengers.mrcNumber"]
"method"="nil"
["serviceItems.passengers.clientInformations.fidNumber"]
"method"="set"
"args"=["xxx"]
["serviceItems.passengers.clientCard.cardNumber"]
"method"="set"
"args"=["xxx"]
["serviceItems.passengers.sncfAgentAdditionalProperties.agentId"]
"method"="set"
"args"=["xxx"]
["serviceItems.railTransportationContracts.holder.firstName"]
"method"="set"
"args"=["prenom"]
["serviceItems.railTransportationContracts.holder.lastName"]
"method"="set"
"args"=["nom"]
["serviceItems.vehicleOwner.iuc"]
"method"="set"
"args"=["xxx"]
["serviceItems.vehicleOwner.firstName"]
"method"="set"
"args"=["prenom"]
["serviceItems.vehicleOwner.lastName"]
"method"="set"
"args"=["nom"]
["serviceItems.vehicles.plateNumber"]
"method"="set"
"args"=["xxx"]

#["serviceItems.passengers.birthDate"]
#"method"="date"
#"args"=["Jan 2 15:04:05 -0700 MST 2006", "Jan 1 00:00:00 -0100 MST 1900"]
#["passengers.birthDate"]
#"method"="date"
#"args"=["Jan 2 15:04:05 -0700 MST 2006", "Jan 1 00:00:00 -0100 MST 1900"]


["contactInformation.emailAddress"]
"method"="set"
"args" = [ "toto@toto.fr" ]
`
	cleaner = doccleaner.NewDocCleaner(strings.NewReader(config))
	document := CreateOrder(5)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		anonymizeDocument(document)
	}

	if document["customers"].([]interface{})[0].(bson.M)["iuc"] != "xxx" {
		b.Errorf("document is not really anonymized:\n %+v\n", document)
	}
}

func TestAnonymizeServiceItem(t *testing.T) {
	// given
	document := CreateServiceItem()
	config := `
["contactInformation.firstname"]
"method"="set"
"args" = [ "firstname" ]
["contactInformation.name"]
"method"="set"
"args" = [ "name" ]
["contactInformation.emailAddress"]
"method"="set"
"args" = [ "toto@toto.fr" ]
["contactInformation.address1"]
"method"="set"
"args" = [ "address1" ]
["contactInformation.address2"]
"method"="set"
"args" = [ "address2" ]
["contactInformation.address3"]
"method"="set"
"args" = [ "address3" ]
["contactInformation.address4"]
"method"="set"
"args" = [ "address4" ]
["contactInformation.city"]
"method"="set"
"args" = [ "city" ]
["contactInformation.zipCode"]
"method"="set"
"args" = [ "zipCode" ]
["contactInformation.country"]
"method"="set"
"args" = [ "country" ]
["contactInformation.mobilePhoneNumber"]
"method"="set"
"args" = [ "0123456789" ]
["contactInformation.landlinePhoneNumbers"]
"method"="nil"
["passengers.emailAddress"]
"method"="set"
"args" = [ "toto@toto.fr" ]
["passengers.iuc"]
"method"="set"
"args" = [ "iuc" ]
["passengers.firstName"]
"method"="set"
"args" = [ "firstName" ]
["passengers.lastName"]
"method"="set"
"args" = [ "lastName" ]
["passengers.mobilePhoneNumber"]
"method"="set"
"args" = ["0123456789" ]
["passengers.fceNumber"]
"method"="set"
"args"=["fceNumber"]
["passengers.mrcNumber"]
"method"="nil"
["passengers.birthDate"]
"method"="date"
"args"=["Jan 2 15:04:05 -0700 MST 2006", "Jan 1 00:00:00 -0100 MST 1900"]
["passengers.clientInformations.fidNumber"]
"method"="set"
"args"=["xxx"]
["passengers.clientCard.cardNumber"]
"method"="set"
"args"=["xxx"]
["passengers.sncfAgentAdditionalProperties.agentId"]
"method"="set"
"args"=["xxx"]
["railTransportationContracts.holder.firstName"]
"method"="set"
"args"=["prenom"]
["railTransportationContracts.holder.lastName"]
"method"="set"
"args"=["nom"]
["vehicleOwner.iuc"]
"method"="set"
"args"=["xxx"]
["vehicleOwner.firstName"]
"method"="set"
"args"=["prenom"]
["vehicleOwner.lastName"]
"method"="set"
"args"=["nom"]
["vehicles.plateNumber"]
"method"="set"
"args"=["xxx"]
`
	cleaner = doccleaner.NewDocCleaner(strings.NewReader(config))

	// when
	anonymizeDocument(document)

	// then
	testServiceItem(document, t)
}

func TestAnonymizeOrder(t *testing.T) {
	// given
	document := CreateOrder(2)
	config := `
["customers.iuc"]
"method"="set"
"args" = ["xxx"]
["customers.firstname"]
"method"="set"
"args" = [ "prenom" ]
["customers.lastname"]
"method"="set"
"args" = [ "nom" ]
["ownerCustomer.iuc"]
"method"="set"
"args" = [ "xxx" ]

["serviceItems.contactInformation.firstname"]
"method"="set"
"args" = [ "firstname" ]
["serviceItems.contactInformation.name"]
"method"="set"
"args" = [ "name" ]
["serviceItems.contactInformation.emailAddress"]
"method"="set"
"args" = [ "toto@toto.fr" ]
["serviceItems.contactInformation.address1"]
"method"="set"
"args" = [ "address1" ]
["serviceItems.contactInformation.address2"]
"method"="set"
"args" = [ "address2" ]
["serviceItems.contactInformation.address3"]
"method"="set"
"args" = [ "address3" ]
["serviceItems.contactInformation.address4"]
"method"="set"
"args" = [ "address4" ]
["serviceItems.contactInformation.city"]
"method"="set"
"args" = [ "city" ]
["serviceItems.contactInformation.zipCode"]
"method"="set"
"args" = [ "zipCode" ]
["serviceItems.contactInformation.country"]
"method"="set"
"args" = [ "country" ]
["serviceItems.contactInformation.mobilePhoneNumber"]
"method"="set"
"args" = [ "0123456789" ]
["serviceItems.contactInformation.landlinePhoneNumbers"]
"method"="nil"
["serviceItems.passengers.emailAddress"]
"method"="set"
"args" = [ "toto@toto.fr" ]
["serviceItems.passengers.iuc"]
"method"="set"
"args" = [ "iuc" ]
["serviceItems.passengers.firstName"]
"method"="set"
"args" = [ "firstName" ]
["serviceItems.passengers.lastName"]
"method"="set"
"args" = [ "lastName" ]
["serviceItems.passengers.mobilePhoneNumber"]
"method"="set"
"args" = ["0123456789" ]
["serviceItems.passengers.fceNumber"]
"method"="set"
"args"=["fceNumber"]
["serviceItems.passengers.mrcNumber"]
"method"="nil"
["serviceItems.passengers.birthDate"]
"method"="date"
"args"=["Jan 2 15:04:05 -0700 MST 2006", "Jan 1 00:00:00 -0100 MST 1900"]
["serviceItems.passengers.clientInformations.fidNumber"]
"method"="set"
"args"=["xxx"]
["serviceItems.passengers.clientCard.cardNumber"]
"method"="set"
"args"=["xxx"]
["serviceItems.passengers.sncfAgentAdditionalProperties.agentId"]
"method"="set"
"args"=["xxx"]
["serviceItems.railTransportationContracts.holder.firstName"]
"method"="set"
"args"=["prenom"]
["serviceItems.railTransportationContracts.holder.lastName"]
"method"="set"
"args"=["nom"]
["serviceItems.vehicleOwner.iuc"]
"method"="set"
"args"=["xxx"]
["serviceItems.vehicleOwner.firstName"]
"method"="set"
"args"=["prenom"]
["serviceItems.vehicleOwner.lastName"]
"method"="set"
"args"=["nom"]
["serviceItems.vehicles.plateNumber"]
"method"="set"
"args"=["xxx"]


["contactInformation.emailAddress"]
"method"="set"
"args" = [ "toto@toto.fr" ]
`
	cleaner = doccleaner.NewDocCleaner(strings.NewReader(config))

	// when
	anonymizeDocument(document)

	// then
	testCustomer(document, t)
	testOwnerCustomer(document, t)
	testServiceItems(document, t, 2)
}
