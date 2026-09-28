package main

import (
    "net/http"
	"fmt"
    "github.com/gin-gonic/gin"
	"errors"
	"io"
	"encoding/json"
	"github.com/echa/code/iata"
	"time"
	"strings"
	"regexp"
	"strconv"
	"sync"
	"github.com/patrickmn/go-cache"
)

type SearchResult struct {
	SearchCriteria SearchCriteria `json:"search_criteria"`
	Metadata SearchResultSearchMetadata `json:"metadata"`
	Flights []SearchResultSearchFlight `json:"flights"`
}

type SearchCriteria struct {
	Origin string `json:"origin"`
	Destination string `json:"destination"`
	DepartureDate string `json:"departure_date"`
	Passengers uint `json:"passengers"`
	CabinClass string `json:"cabin_class"`
}

type SearchResultSearchMetadata struct {
	TotalResults int `json:"total_results"`
	ProvidersQueried int `json:"providers_queried"`
	ProvidersSucceeded int `json:"providers_succeeded"`
	ProvidersFailed int `json:"providers_failed"`
	SearchTimeMs int64 `json:"search_time_ms"`
	CacheHit bool `json:"cache_hit"`
}

type SearchResultSearchFlight struct {
	ID string `json:"id"`
	Provider string `json:"provider"`
	Airline SearchResultSearchFlightAirline `json:"airline"`
	FlightNumber string `json:"flight_number"`
	Departure SearchResultSearchFlightDepartureArrival `json:"departure"`
	Arrival SearchResultSearchFlightDepartureArrival `json:"arrival"`
	Duration SearchResultSearchFlightDuration `json:"duration"`
	Stops int `json:"stops"`
	Price SearchResultSearchFlightPrice `json:"price"`
	AvailableSeats uint `json:"available_seats"`
	CabinClass string `json:"cabin_class"`
	Aircraft *string `json:"aircraft"`
	Amenities []string `json:"amenities"`
	Baggage SearchResultSearchFlightBaggage `json:"baggage"`
}

type SearchResultSearchFlightAirline struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type SearchResultSearchFlightDepartureArrival struct {
	Airport string `"json:"airport"`
	City string `json:"city"`
	Datetime string `json:"datetime"`
	Timestamp int64 `json:"timestamp`
}

type SearchResultSearchFlightDuration struct {
	TotalMinutes int64 `json:"total_minutes"`
	Formatted string `json:"formatted"`
}

type SearchResultSearchFlightPrice struct {
	Amount uint `json:"amount"`
	Currency string `json:"currency"`
}

type SearchResultSearchFlightBaggage struct {
	CarryOn string `json:"carry_on"`
	Checked string `json:"checked"`
}

type AirasiaPayload struct {
	Status string `json:"status"`
	Flights []AirasiaFlight `json:"flights"`
}

// Airasia related
type AirasiaFlight struct {
	FlightCode string `json:"flight_code"`
	Airline string `json:"airline"`
	FromAirport string `json:"from_airport"`
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

func searchAirAsia(searchCriteria SearchCriteria) ([]SearchResultSearchFlight, error) {
	resp, err := http.Get("http://localhost:8080/airasia/search")
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, errors.New("Error on getting air asia")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("Error on reading body")
	}

	var airAsiaPayload AirasiaPayload

	err = json.Unmarshal(body, &airAsiaPayload)
	if err != nil {
		return nil, err
	}
	
	if (airAsiaPayload.Status != "ok") {
		return nil, errors.New("Error on airasia status not ok")
	}

	var result = []SearchResultSearchFlight {}
	
	for _, flight := range airAsiaPayload.Flights {
		// getting departure and arrival timestamp
		departureTime, err := time.Parse(time.RFC3339, flight.DepartTime)
		if err != nil {
			fmt.Println("Error on parsing departure time for airasia")
		}

		departureTimestamp := departureTime.Unix()

		arrivalTime, err := time.Parse(time.RFC3339, flight.ArriveTime)
		if err != nil {
			fmt.Println("Error on parsing arrival time for airasia")
		}

		arrivalTimestamp := arrivalTime.Unix()

		// Getting durations from arrive time and depart time
		totalMinutes := (arrivalTimestamp - departureTimestamp) / 60

		hours := totalMinutes / 60
		remainingMinutes := totalMinutes % 60

		formattedDuration := fmt.Sprintf("%dh %dm", hours, remainingMinutes)

		// Baggage
		var carryOn string
		if strings.Contains(flight.BaggageNote, "Cabin baggage only") {
			carryOn = "Cabin baggage only"
		} else {
			carryOn = "undefined"
		}

		var checked string
		if strings.Contains(flight.BaggageNote, "checked bags additional fee") {
			checked = "Additional fee"
		} else {
			checked = "undefined"
		}

		flightItem := SearchResultSearchFlight {
			ID: flight.FlightCode + "_" + flight.Airline,
			Provider: flight.Airline,
			Airline: SearchResultSearchFlightAirline {
				Name: flight.Airline,
				Code: "QZ",
			},
			FlightNumber: flight.FlightCode,
			Departure: SearchResultSearchFlightDepartureArrival {
				Airport: flight.FromAirport,
				City: iata.ParseAirportCode(flight.FromAirport).Airport().Name,
				Datetime: flight.DepartTime,
				Timestamp: departureTimestamp,
			},
			Arrival: SearchResultSearchFlightDepartureArrival {
				Airport: flight.ToAirport,
				City: iata.ParseAirportCode(flight.ToAirport).Airport().Name,
				Datetime: flight.ArriveTime,
				Timestamp: arrivalTimestamp,
			},
			Duration: SearchResultSearchFlightDuration {
				TotalMinutes: totalMinutes,
				Formatted: formattedDuration,
			},
			Stops: len(flight.Stops),
			Price: SearchResultSearchFlightPrice {
				Amount: flight.PriceIdr,
				Currency: "IDR",
			},
			AvailableSeats: flight.Seats,
			CabinClass: flight.CabinClass,
			Aircraft: nil,
			Amenities: []string{},
			Baggage: SearchResultSearchFlightBaggage {
				CarryOn: carryOn,
				Checked: checked,
			},
		}

		result = append(result, flightItem)
	}
	
	// Do filter in here
	var filteredFlight = filterProviderAirlineResult(searchCriteria, result)

	return filteredFlight, nil
}

// Batik air related

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


func searchBatikAir(searchCriteria SearchCriteria) ([]SearchResultSearchFlight, error) {
	resp, err := http.Get("http://localhost:8080/batikAir/search")
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, errors.New("Error on getting batik air")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("Error on reading body")
	}

	var batikAirPayload BatikAirPayload

	err = json.Unmarshal(body, &batikAirPayload)
		if err != nil {
		return nil, err
	}
	
	if (batikAirPayload.Code != 200) {
		return nil, errors.New("Error on batik air status not 200")
	}

	var result = []SearchResultSearchFlight {}

	for _, flight := range batikAirPayload.Results {
		// getting departure and arrival timestamp
		departureTime, err := time.Parse("2006-01-02T15:04:05-0700", flight.DepartureDateTime)
		if err != nil {
			fmt.Println("Error on parsing departure time for batik air")
		}

		departureTimestamp := departureTime.Unix()

		arrivalTime, err := time.Parse("2006-01-02T15:04:05-0700", flight.ArrivalDateTime)
		if err != nil {
			fmt.Println("Error on parsing arrival time for airasia")
		}

		arrivalTimestamp := arrivalTime.Unix()

		// Getting durations from arrive time and depart time
		totalMinutes := (arrivalTimestamp - departureTimestamp) / 60

		hours := totalMinutes / 60
		remainingMinutes := totalMinutes % 60

		formattedDuration := fmt.Sprintf("%dh %dm", hours, remainingMinutes)

		// Cabin class
		var cabinClass string
		if (flight.Fare.Class == "Y") {
			cabinClass = "economy"
		} else {
			cabinClass = "undefined"
		}
		
		// // Baggage
		var carryOn string
		re := regexp.MustCompile(`(?i)\d+\s*kg\s+cabin`)
		cabinText := re.FindString(flight.BaggageInfo)

		if (cabinText != "") {
			carryOn = cabinText
		} else {
			carryOn = "undefined"
		}

		// var checked string
		var checked string
		re2 := regexp.MustCompile(`(?i)\d+\s*kg\s+checked`)
		checkedText := re2.FindString(flight.BaggageInfo)

		if (checkedText != "") {
			checked = checkedText
		} else {
			checked = "undefined"
		}

		flightItem := SearchResultSearchFlight {
			ID: flight.FlightNumber + "_" + flight.AirlineName,
			Provider: flight.AirlineName,
			Airline: SearchResultSearchFlightAirline {
				Name: flight.AirlineName,
				Code: flight.AirlineIATA,
			},
			FlightNumber: flight.FlightNumber,
			Departure: SearchResultSearchFlightDepartureArrival {
				Airport: flight.Origin,
				City: iata.ParseAirportCode(flight.Origin).Airport().Name,
				Datetime: departureTime.Format(time.RFC3339),
				Timestamp: departureTimestamp,
			},
			Arrival: SearchResultSearchFlightDepartureArrival {
				Airport: flight.Destination,
				City: iata.ParseAirportCode(flight.Destination).Airport().Name,
				Datetime: arrivalTime.Format(time.RFC3339),
				Timestamp: arrivalTimestamp,
			},
			Duration: SearchResultSearchFlightDuration {
				TotalMinutes: totalMinutes,
				Formatted: formattedDuration,
			},
			Stops: flight.NumberOfStops,
			Price: SearchResultSearchFlightPrice {
				Amount: flight.Fare.TotalPrice,
				Currency: flight.Fare.CurrencyCode,
			},
			AvailableSeats: flight.SeatsAvailable,
			CabinClass: cabinClass,
			Aircraft: &flight.AircraftModel,
			Amenities: flight.OnboardServices,
			Baggage: SearchResultSearchFlightBaggage {
				CarryOn: carryOn,
				Checked: checked,
			},
		}

		result = append(result, flightItem)
	}

	// Do filter in here
	var filteredFlight = filterProviderAirlineResult(searchCriteria, result)

	return filteredFlight, nil
}

// Garuda Indonesia related
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

func searchGarudaIndonesia(searchCriteria SearchCriteria) ([]SearchResultSearchFlight, error) {
	resp, err := http.Get("http://localhost:8080/garudaIndonesia/search")
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, errors.New("Error on getting Garuda Indonesia")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("Error on reading body")
	}

	var garudaIndonesiaPayload GarudaIndonesiaPayload

	err = json.Unmarshal(body, &garudaIndonesiaPayload)
		if err != nil {
		return nil, err
	}
	
	if (garudaIndonesiaPayload.Status != "success") {
		return nil, errors.New("Error on Garuda Indonesia status not success")
	}

	var result = []SearchResultSearchFlight {}

	for _, flight := range garudaIndonesiaPayload.Flights {
		// getting departure and arrival timestamp
		departureTime, err := time.Parse(time.RFC3339, flight.Departure.Time)
		if err != nil {
			fmt.Println("Error on parsing departure time for garuda air", err)
		}

		departureTimestamp := departureTime.Unix()

		arrivalTime, err := time.Parse(time.RFC3339, flight.Arrival.Time)
		if err != nil {
			fmt.Println("Error on parsing arrival time for Garuda Indonesia")
		}

		arrivalTimestamp := arrivalTime.Unix()

		// Getting durations from arrive time and depart time
		totalMinutes := (arrivalTimestamp - departureTimestamp) / 60

		hours := totalMinutes / 60
		remainingMinutes := totalMinutes % 60

		formattedDuration := fmt.Sprintf("%dh %dm", hours, remainingMinutes)

	// 	// Cabin class
	// 	var cabinClass string
	// 	if (flight.Fare.Class == "Y") {
	// 		cabinClass = "economy"
	// 	} else {
	// 		cabinClass = "undefined"
	// 	}
		
	// 	// // Baggage
	// 	var carryOn string
	// 	re := regexp.MustCompile(`(?i)\d+\s*kg\s+cabin`)
	// 	cabinText := re.FindString(flight.BaggageInfo)

	// 	if (cabinText != "") {
	// 		carryOn = cabinText
	// 	} else {
	// 		carryOn = "undefined"
	// 	}

	// 	// var checked string
	// 	var checked string
	// 	re2 := regexp.MustCompile(`(?i)\d+\s*kg\s+checked`)
	// 	checkedText := re2.FindString(flight.BaggageInfo)

	// 	if (checkedText != "") {
	// 		checked = checkedText
	// 	} else {
	// 		checked = "undefined"
	// 	}

		flightItem := SearchResultSearchFlight {
			ID: flight.FlightID + "_" + flight.Airline,
			Provider: flight.Airline,
			Airline: SearchResultSearchFlightAirline {
				Name: flight.Airline,
				Code: flight.AirlineCode,
			},
			FlightNumber: flight.FlightID,
			Departure: SearchResultSearchFlightDepartureArrival {
				Airport: flight.Departure.Airport,
				City: flight.Departure.City,
				Datetime: flight.Departure.Time,
				Timestamp: departureTimestamp,
			},
			Arrival: SearchResultSearchFlightDepartureArrival {
				Airport: flight.Arrival.Airport,
				City: flight.Arrival.City,
				Datetime: flight.Arrival.Time,
				Timestamp: arrivalTimestamp,
			},
			Duration: SearchResultSearchFlightDuration {
				TotalMinutes: totalMinutes,
				Formatted: formattedDuration,
			},
			Stops: flight.Stops,
			Price: SearchResultSearchFlightPrice {
				Amount: flight.Price.Amount,
				Currency: flight.Price.Currency,
			},
			AvailableSeats: flight.AvailableSeats,
			CabinClass: flight.FareClass,
			Aircraft: &flight.Aircraft,
			Amenities: flight.Amenities,
			Baggage: SearchResultSearchFlightBaggage {
				CarryOn: strconv.FormatUint(uint64(flight.Baggage.CarryOn), 10),
				Checked: strconv.FormatUint(uint64(flight.Baggage.Checked), 10),
			},
		}

		result = append(result, flightItem)
	}

	// Do filter in here
	var filteredFlight = filterProviderAirlineResult(searchCriteria, result)

	return filteredFlight, nil
}

// Lion air
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

func searchLionAir(searchCriteria SearchCriteria) ([]SearchResultSearchFlight, error) {
	resp, err := http.Get("http://localhost:8080/lionAir/search")
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, errors.New("Error on getting Lion air")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("Error on reading body")
	}

	var lionAirPayload LionAirPayload

	err = json.Unmarshal(body, &lionAirPayload)
		if err != nil {
		return nil, err
	}
	
	if (lionAirPayload.Success != true) {
		return nil, errors.New("Error on batik air status not 200")
	}

	var result = []SearchResultSearchFlight {}

	for _, flight := range lionAirPayload.Data.AvailableFlights {
		// getting departure and arrival timestamp
		locationDeparture, err := time.LoadLocation(flight.Schedule.DepartureTimezone)
		if err != nil {
			fmt.Println("Error on parsing location from timezone for lion air")
		}


		departureTime, err := time.ParseInLocation(
			"2006-01-02T15:04:05", 
			flight.Schedule.Departure,
			locationDeparture,
		)
		if err != nil {
			fmt.Println("Error on parsing departure time for lion air")
		}

		departureTimestamp := departureTime.Unix()

		locationArrival, err := time.LoadLocation(flight.Schedule.ArrivalTimezone)
		if err != nil {
			fmt.Println("Error on parsing location from timezone for lion air")
		}


		arrivalTime, err := time.ParseInLocation(
			"2006-01-02T15:04:05", 
			flight.Schedule.Arrival,
			locationArrival,
		)
		if err != nil {
			fmt.Println("Error on parsing Arrival time for lion air")
		}

		arrivalTimestamp := arrivalTime.Unix()

		// Getting durations from arrive time and depart time
		totalMinutes := (arrivalTimestamp - departureTimestamp) / 60

		hours := totalMinutes / 60
		remainingMinutes := totalMinutes % 60

		formattedDuration := fmt.Sprintf("%dh %dm", hours, remainingMinutes)

		// Amenities
		amenities := []string{}
		if (flight.Services.WifiAvailable) {
			amenities = append(amenities, "wifi")
		}

		if (flight.Services.MealsIncluded) {
			amenities = append(amenities, "meal")
		}

		// // Cabin class
		var cabinClass string
		if (flight.Pricing.FareType == "ECONOMY") {
			cabinClass = "economy"
		} else {
			cabinClass = "undefined"
		}
		
		// // // Baggage
		// var carryOn string
		// re := regexp.MustCompile(`(?i)\d+\s*kg\s+cabin`)
		// cabinText := re.FindString(flight.BaggageInfo)

		// if (cabinText != "") {
		// 	carryOn = cabinText
		// } else {
		// 	carryOn = "undefined"
		// }

		// // var checked string
		// var checked string
		// re2 := regexp.MustCompile(`(?i)\d+\s*kg\s+checked`)
		// checkedText := re2.FindString(flight.BaggageInfo)

		// if (checkedText != "") {
		// 	checked = checkedText
		// } else {
		// 	checked = "undefined"
		// }

		flightItem := SearchResultSearchFlight {
			ID: flight.ID + "_" + flight.Carrier.Name,
			Provider: flight.Carrier.Name,
			Airline: SearchResultSearchFlightAirline {
				Name: flight.Carrier.Name,
				Code: flight.Carrier.Iata,
			},
			FlightNumber: flight.ID,
			Departure: SearchResultSearchFlightDepartureArrival {
				Airport: flight.Route.From.Code,
				City: flight.Route.From.City,
				Datetime: departureTime.Format(time.RFC3339),
				Timestamp: departureTimestamp,
			},
			Arrival: SearchResultSearchFlightDepartureArrival {
				Airport: flight.Route.To.Code,
				City: flight.Route.To.City,
				Datetime: arrivalTime.Format(time.RFC3339),
				Timestamp: arrivalTimestamp,
			},
			Duration: SearchResultSearchFlightDuration {
				TotalMinutes: totalMinutes,
				Formatted: formattedDuration,
			},
			Stops: flight.StopCount,
			Price: SearchResultSearchFlightPrice {
				Amount: flight.Pricing.Total,
				Currency: flight.Pricing.Currency,
			},
			AvailableSeats: flight.SeatsLeft,
			CabinClass: cabinClass,
			Aircraft: &flight.PlaneType,
			Amenities: amenities,
			Baggage: SearchResultSearchFlightBaggage {
				CarryOn: flight.Services.BaggageAllowance.Cabin,
				Checked: flight.Services.BaggageAllowance.Hold,
			},
		}

		result = append(result, flightItem)
	}

	// Do filter in here
	var filteredFlight = filterProviderAirlineResult(searchCriteria, result)

	return filteredFlight, nil
}

func filterProviderAirlineResult(searchCriteria SearchCriteria, flights []SearchResultSearchFlight) []SearchResultSearchFlight{
	// Do filter in here
	var filteredFlight []SearchResultSearchFlight
	for _,flight := range flights {
		// convert date time
		specDate := searchCriteria.DepartureDate
		flightDate := flight.Departure.Datetime

		t1, err := time.Parse("2006-01-02", specDate)
		if err != nil {
			fmt.Println("error on parsing time for spec date", specDate)
			t1 = time.Now()
		}

		t2, err := time.Parse(time.RFC3339, flightDate)
		if err != nil {
			fmt.Println("error on parsing time for flight date", flightDate)
			t2 = time.Now().Add(-10000 * time.Second)
		}

		sameDate := t1.Year() == t2.Year() &&
		t1.Month() == t2.Month() &&
		t1.Day() == t2.Day()

		if (searchCriteria.Origin == flight.Departure.Airport && 
			searchCriteria.Destination == flight.Arrival.Airport &&
			sameDate &&
			searchCriteria.Passengers <= flight.AvailableSeats &&
			searchCriteria.CabinClass == flight.CabinClass) {
			filteredFlight = append(filteredFlight, flight)
		}
	}

	return filteredFlight
}

func searchAirlineAggregator(c *gin.Context) {
	start := time.Now()

	// Construct Search specs
	passengers, err := strconv.ParseUint(c.Query("passengers"), 10, 0)
	if err != nil {
		// handle invalid passengers parameter
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid passengers",
		})
		return
	}

	searchCriteria := SearchCriteria {
		Origin: c.Query("origin"),
		Destination: c.Query("destination"),
		DepartureDate: c.Query("departure_date"),
		Passengers: uint(passengers),
		CabinClass: c.Query("cabin_class"),
	}


	// Construct search result
	var searchResult SearchResult

	var cacheKey = fmt.Sprintf("%v",searchCriteria)

	if cached, found := searchAirlineAggregatorCache.Get(cacheKey); found {
		searchResult = cached.(SearchResult)

		// update metadata
		duration := time.Since(start).Milliseconds()
		searchResult.Metadata.SearchTimeMs = duration
		searchResult.Metadata.CacheHit = true
		
		
		c.IndentedJSON(http.StatusOK, searchResult)
		return
	}

	searchResult.SearchCriteria = searchCriteria

	var providersQueried = 0
	var providersError = 0

	// Async version
	searchers := []func(SearchCriteria) ([]SearchResultSearchFlight, error){
		searchAirAsia,
		searchBatikAir,
		searchGarudaIndonesia,
		searchLionAir,
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	providersQueried = len (searchers)
	providersError = 0

	for _,search := range searchers {
		wg.Add(1)

		go func(search func(SearchCriteria) ([]SearchResultSearchFlight, error)) {
			defer wg.Done()

			flights, err := search(searchCriteria)
			if (err != nil) {
				mu.Lock()
				providersError++
				mu.Unlock()
				return
			}

			mu.Lock()
			searchResult.Flights = append(searchResult.Flights, flights...)
			mu.Unlock()
		}(search)
	}

	wg.Wait()

	// Sync version
	// // Search air asia
	// var searchAirAsiaFlightResult, err = searchAirAsia()
	// providersQueried += 1
	// if (err != nil) {
	// 	providersError += 1
	// } else {
	// 	// add to flight
	// 	searchResult.Flights = append(searchResult.Flights, searchAirAsiaFlightResult...)
	// }

	// // Search batik air
	// var searchBatikAirFlightResult, errBatikAir = searchBatikAir()
	// providersQueried += 1
	// if (errBatikAir != nil) {
	// 	providersError += 1
	// } else {
	// 	// add to flight
	// 	searchResult.Flights = append(searchResult.Flights, searchBatikAirFlightResult...)
	// }

	// // Search Garuda Indonesia
	// var searchGarudaIndonesiaFlightResult, errGarudaIndonesia = searchGarudaIndonesia()
	// providersQueried += 1
	// if (errGarudaIndonesia != nil) {
	// 	providersError += 1
	// } else {
	// 	// add to flight
	// 	searchResult.Flights = append(searchResult.Flights, searchGarudaIndonesiaFlightResult...)
	// }

	// // Search Lion Air
	// var searchLionAirFlightResult, errLionAir = searchLionAir()
	// providersQueried += 1
	// if (errLionAir != nil) {
	// 	providersError += 1
	// } else {
	// 	// add to flight
	// 	searchResult.Flights = append(searchResult.Flights, searchLionAirFlightResult...)
	// }

	duration := time.Since(start).Milliseconds()
	
	// aggregate search results
	searchResult.Metadata.TotalResults = len(searchResult.Flights)
	searchResult.Metadata.ProvidersQueried = providersQueried
	searchResult.Metadata.ProvidersSucceeded = providersQueried - providersError
	searchResult.Metadata.ProvidersFailed = providersError
	searchResult.Metadata.SearchTimeMs = duration

	searchAirlineAggregatorCache.Set(cacheKey, searchResult, cache.DefaultExpiration)
	
	c.IndentedJSON(http.StatusOK, searchResult)
}

var searchAirlineAggregatorCache = cache.New(
	5*time.Second, // default expiration
    10*time.Second, // cleanup interval
)

func main() {
    router := gin.Default()
	router.GET("airlineAggregator/search", searchAirlineAggregator)

	router.Run("localhost:8081")
}