There are 2 services in this: 
1. external-web-service for xxx a
2. airline-aggregator for xxx.

To start External Web Services
```
cd external-web-service
go get .
go run .
```

To test External Web Services
```
curl http://localhost:8080/airasia/search
curl http://localhost:8080/batikAir/search
curl http://localhost:8080/garudaIndonesia/search
curl http://localhost:8080/lionAir/search
```

To start Airline Aggregator
```
cd airline-aggregator
go get .
go run .
```
To test airline aggregator
```
curl "localhost:8081/airlineAggregator/search?origin=CGK&destination=DPS&departure_date=2025-12-15&passengers=1&cabin_class=economy"
```

Design Choices
Filter from provider levels. because there may be provider that provide it which will make time lesser
using "github.com/patrickmn/go-cache" for simpler use of cache. May need to use other caches in case the case becomes more complex