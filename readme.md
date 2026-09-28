# airline_aggregation
Simple airline aggregation from providers to test golang.
## How To Start Services
There are 2 services in this project:
1. external-web-service for simulating external provider for airline.
2. airline-aggregator for aggregating flights data from external-web-service api above.

To start External Web Services
```
# start new terminal session

cd external-web-service
go get .
go run .
```

To start Airline Aggregator
```
# start new terminal session

cd airline-aggregator
go get .
go run .
```

## How To Test Services

### Test External Web Services
```
# start new terminal session

curl http://localhost:8080/airasia/search
curl http://localhost:8080/batikAir/search
curl http://localhost:8080/garudaIndonesia/search
curl http://localhost:8080/lionAir/search
```

### Test airline aggregator
```
# start new terminal session

curl "localhost:8081/airlineAggregator/search?origin=CGK&destination=DPS&departure_date=2025-12-15&passengers=1&cabin_class=economy&sort_criteria=price_highest"
```
sample options for origin CGK DPS<br>
sample options for destination CGK DPS<br>
sample options for departure_date 2025-12-15 2026-05-15<br>
sample options for passengers 1 20 100<br>
options for cabin_class economy<br>
options for sort_criteria price_highest price_lowest duration_shortest duration_longest<br>

## Design Choices
1. I separate airline provider service to simulate the different kinds of condition of the providers. e.g. on Air Asia has 50-150 ms latency with 90% success rate.
1. Every provider API caller has each separate functions because of different payloads. But the result will be converted to type `SearchResultSearchFlight` and merge as one results. 
1. The call to providers are using goroutine to make the call asynchronous and reduce latency when calling the aggregator service. 
1. Cache is using "github.com/patrickmn/go-cache" for simpler implementation. In case for larger throughput, may need to use other cache library.
1. Filter is done on provider levels because there may be provider that provide filter capabilities which will make filter done by their side. (currently done on our side because mock api doesn't provide the capability.).