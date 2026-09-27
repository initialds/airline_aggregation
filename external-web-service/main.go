package main

import (
    "net/http"
    "github.com/gin-gonic/gin"
	"os"
	"fmt"
	"io"
	"encoding/json"
	"time"
	"math/rand"
)

type AirasiaPayload struct {
	Status string `json:"status"`
	Flights []AirasiaFlight `json:"flights"`
}

type AirasiaFlight struct {
	FlightCode string `json:"flight_code"`
	Airline string `json:"airline"`
	FromAiport string `json:"from_airport"`
	ToAirport string `json:"to_airport"`
	DepartTime string `json:"depart_time"`
	ArriveTime string `json:"arrive_time"`
	DurationHours float64 `json:"duration_hours"`
	DirectFlight bool `json:"direct_flight"`
	Stops []AirasiaFlightStops `json:"stops"`
	PriceIdr uint `json:"price_idr"`
	Seats uint `json:"seats"`
	CabinClass string `json:"cabin_class"`
	BaggageNote string `json:"baggage_note"`
}

type AirasiaFlightStops struct {
	Airport string `json:"airport"`
	WaitTimeMinutes uint `json:"wait_time_minutes"`
}

func getAirasiaResponseJson() AirasiaPayload {
    // Open our jsonFile
    jsonFile, err := os.Open("payloads/airasia_search_response.json")
    // if we os.Open returns an error then handle it
    if err != nil {
        println(err)
    }

    // defer the closing of our jsonFile so that we can parse it later on
    defer jsonFile.Close()

	// read our opened jsonFile as a byte array.
    byteValue, err := io.ReadAll(jsonFile)
    if err != nil {
        fmt.Println("Error reading file:", err)
        return AirasiaPayload {
			Status: "failed",
		}
    }

	
	
	var airasiaPayload AirasiaPayload
	
	err = json.Unmarshal(byteValue, &airasiaPayload)
    if err != nil {
        fmt.Println("Error unmarshalling JSON:", err)
        return AirasiaPayload {
			Status: "failed",
		}
    }

	return airasiaPayload
}

var airasiaPayload = getAirasiaResponseJson()


func getAirasiaPayload(c *gin.Context) {
	// 10 % simulate error
	// 90 % simulate delay 50-150 ms
	var errorPercentage = 10
	if (rand.Intn(errorPercentage) == 0) {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	
	var max = 150
	var min = 50
	var delay = rand.Intn(max-min) + min
	time.Sleep(time.Duration(delay) * time.Millisecond)

	c.IndentedJSON(http.StatusOK, airasiaPayload)
}

type BatikAirPayload struct {
	Code uint `json:"code"`
	Message string `json:"message"`
	Results []BatikAirFlight `json:"results"`
}

type BatikAirFlight struct {
	FlightNumber string `json:"flightNumber"`
	AirlineName string `json:"airlineName"`
	AirlineIATA string `json:"airlineIATA"`
	Origin string `json:"origin"`
	Destination string `json:"destination"`
	DepartureDateTime string `json:"departureDateTime"`
	ArrivalDateTime string `json:"arrivalDateTime"`
	TravelTime string `json:"travelTime"`
	NumberOfStops int `json:numberOfStops`
	Connections []BatikAirFlightConnections `json:"connections"`
	Fare BatikAirFlightFare `json:"fare"`
	SeatsAvailable uint `json:"seatsAvailable"`
	AircraftModel string `json:"aircraftModel"`
	BaggageInfo string `json:"baggageInfo"`
	OnboardServices []string `json:"onboardServices"`
}

type BatikAirFlightFare struct {
	BasePrice uint `json:"basePrice"`
	Taxes uint `json:"taxes"`
	TotalPrice uint `json:"totalPrice`
	CurrencyCode string `json:"currencyCode`
	Class string `json:"class"`
}

type BatikAirFlightConnections struct {
	StopAirport string `json:"stopAirport"`
	StopDuration string `json:"stopDuration"`
}

func getBatikAirResponseJson() BatikAirPayload {
    // Open our jsonFile
    jsonFile, err := os.Open("payloads/batik_air_search_response.json")
    // if we os.Open returns an error then handle it
    if err != nil {
        println(err)
    }

    // defer the closing of our jsonFile so that we can parse it later on
    defer jsonFile.Close()

	// read our opened jsonFile as a byte array.
    byteValue, err := io.ReadAll(jsonFile)
    if err != nil {
        fmt.Println("Error reading file:", err)
        return BatikAirPayload {
			Code: 500,
			Message:"ERROR",
		}
    }

	
	
	var batikAirPayload BatikAirPayload
	
	err = json.Unmarshal(byteValue, &batikAirPayload)
    if err != nil {
        fmt.Println("Error unmarshalling JSON:", err)
        return BatikAirPayload {
			Code: 500,
			Message:"ERROR",
		}
    }

	return batikAirPayload
}

var batikAirPayload = getBatikAirResponseJson()


func getBatikAirPayload(c *gin.Context) {
	var max = 400
	var min = 200
	var delay = rand.Intn(max-min) + min
	time.Sleep(time.Duration(delay) * time.Millisecond)

	c.IndentedJSON(http.StatusOK, batikAirPayload)
}

type GarudaIndonesiaPayload struct {
	Status string `json:"status"`
	Flights []GarudaIndonesiaFlight `json:"flights"`
}

type GarudaIndonesiaFlight struct {
	FlightID string `json:"flight_id"`
	Airline string `json:"airline"`
	AirlineCode string `json:"airline_code"`
	Departure GarudaIndonesiaFlightDepartureArrival `json:"departure"`
	Arrival GarudaIndonesiaFlightDepartureArrival `json:"arrival"`
	DurationMinutes uint `json:"duration_minutes"`
	Stops int `json:"stops"`
	Aircraft string `json:"aircraft"`
	Price GarudaIndonesiaFlightPrice `json:"price"`
	Segments []GarudaIndonesiaFlightSegments `json:"segments"`
	AvailableSeats uint `json:"available_seats"`
	FareClass string `json:"fare_class"`
	Baggage GarudaIndonesiaFlightBaggage `json:"baggage"`
	Amenities []string `json:"amenities"`
}

type GarudaIndonesiaFlightDepartureArrival struct {
	Airport string `json:"airport"`
	City string `json:"city"`
	Time string `json"time"`
	Terminal string `json:"terminal"`
}

type GarudaIndonesiaFlightPrice struct {
	Amount uint `json:"amount"`
	Currency string `json:"currency"`
}

type GarudaIndonesiaFlightSegments struct {
	FlightNumber string `json:"flight_number"`
	Departure GarudaIndonesiaFlightSegmentsDepartureArrival `json:"departure"`
	Arrival GarudaIndonesiaFlightSegmentsDepartureArrival `json:"arrival"`
	DurationMinutes uint `json:"duration_minutes"`
}

type GarudaIndonesiaFlightSegmentsDepartureArrival struct {
	Airport string `json:"airport"`
	Time string `json:"time"`
}

type GarudaIndonesiaFlightBaggage struct {
	CarryOn uint `json:"carry_on"`
	Checked uint `json:"checked"`
}

func getGarudaIndonesiaResponseJson() GarudaIndonesiaPayload {
    // Open our jsonFile
    jsonFile, err := os.Open("payloads/garuda_indonesia_search_response.json")
    // if we os.Open returns an error then handle it
    if err != nil {
        println(err)
    }

    // defer the closing of our jsonFile so that we can parse it later on
    defer jsonFile.Close()

	// read our opened jsonFile as a byte array.
    byteValue, err := io.ReadAll(jsonFile)
    if err != nil {
        fmt.Println("Error reading file:", err)
        return GarudaIndonesiaPayload {
			Status: "error",
		}
    }
	
	var garudaIndonesiaPayload GarudaIndonesiaPayload
	
	err = json.Unmarshal(byteValue, &garudaIndonesiaPayload)
    if err != nil {
        fmt.Println("Error unmarshalling JSON:", err)
        return GarudaIndonesiaPayload {
			Status: "error",
		}
    }

	return garudaIndonesiaPayload
}

var garudaIndonesiaPayload = getGarudaIndonesiaResponseJson()


func getGarudaIndonesiaPayload(c *gin.Context) {
	var max = 100
	var min = 50
	var delay = rand.Intn(max-min) + min
	time.Sleep(time.Duration(delay) * time.Millisecond)

	c.IndentedJSON(http.StatusOK, garudaIndonesiaPayload)
}

type LionAirPayload struct {
	Success bool `json:"success"`
	Data LionAirPayloadData `json:"data"`
}

type LionAirPayloadData struct {
	AvailableFlights []LionAirFlight `json:"available_flights"`
}

type LionAirFlight struct {
	ID string `json:"id"`
	Carrier LionAirFlightCarrier `json:"carrier"`
	Route LionAirFlightRoute `json:"route"`
	Schedule LionAirFlightSchedule `json:"schedule"`
	FlightTime uint `json:"flight_time"`
	IsDirect bool `json:"is_direct"`
	StopCount int `json:"stop_count"`
	Layovers []LionAirFlightLayover `json:"layovers"`
	Pricing LionAirFlightPricing `json:"pricing"`
	SeatsLeft uint `json:"seats_left"`
	PlaneType string `json:"plane_type"`
	Services LionAirFlightServices `json:"services"`
}

type LionAirFlightCarrier struct {
	Name string `json:"name"`
	Iata string `json:"iata"`
}

type LionAirFlightRoute struct {
	From LionAirFlightRouteDetail `json:"from"`
	To LionAirFlightRouteDetail `json:"to"`
}

type LionAirFlightRouteDetail struct {
	Code string `json:"code"`
	Name string `json:"name"`
	City string `json:"city"`
}

type LionAirFlightSchedule struct {
	Departure string `json:"departure"`
	DepartureTimezone string `json:"departure_timezone"`
	Arrival string `json:"arrival"`
	ArrivalTimezone string `json:"arrival_timezone"`
}

type LionAirFlightLayover struct {
	Airport string `json:"airport"`
	DurationMinutes uint `json:"duration_minutes"`
}

type LionAirFlightPricing struct {
	Total uint `json:"total"`
	Currency string `json:"currency"`
	FareType string `json:"fare_type"`
}

type LionAirFlightServices struct {
	WifiAvailable bool `json:"wifi_available"`
	MealsIncluded bool `json:"meals_included"`
	BaggageAllowance LionAirFlightServicesBaggageAllowance `json:"baggage_allowance"`
}

type LionAirFlightServicesBaggageAllowance struct {
	Cabin string `json:"cabin"`
	Hold string `json:"hold"`
}



func getLionAirResponseJson() LionAirPayload {
    // Open our jsonFile
    jsonFile, err := os.Open("payloads/lion_air_search_response.json")
    // if we os.Open returns an error then handle it
    if err != nil {
        println(err)
    }

    // defer the closing of our jsonFile so that we can parse it later on
    defer jsonFile.Close()

	// read our opened jsonFile as a byte array.
    byteValue, err := io.ReadAll(jsonFile)
    if err != nil {
        fmt.Println("Error reading file:", err)
        return LionAirPayload {
			Success: false,
		}
    }
	
	var lionAirPayload LionAirPayload
	
	err = json.Unmarshal(byteValue, &lionAirPayload)
    if err != nil {
        fmt.Println("Error unmarshalling JSON:", err)
		return LionAirPayload {
			Success: false,
		}
    }

	return lionAirPayload
}

var lionAirPayload = getLionAirResponseJson()


func getLionAirPayload(c *gin.Context) {
	var max = 200
	var min = 100
	var delay = rand.Intn(max-min) + min
	time.Sleep(time.Duration(delay) * time.Millisecond)

	c.IndentedJSON(http.StatusOK, lionAirPayload)
}

func main() {
    router := gin.Default()
    router.GET("/airasia/search", getAirasiaPayload)
	router.GET("/batikAir/search", getBatikAirPayload)
	router.GET("/garudaIndonesia/search", getGarudaIndonesiaPayload)
	router.GET("/lionAir/search", getLionAirPayload)

    router.Run("localhost:8080")
}