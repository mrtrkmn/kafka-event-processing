# kafka-event-processing

SS2022- Event Processing Project 

## Main Architecture 

<img width="889" alt="Screen Shot 2022-07-26 at 00 59 29" src="https://user-images.githubusercontent.com/13614433/180888666-46067245-ef47-43cb-88ae-a78889018ae3.png">

*(Figure taken from: https://docs.ksqldb.io/en/latest/operate-and-deploy/how-it-works/ )

## Docker compose 

The [docker-compose.yaml](./docker-compose.yml) file is fetched from Confluent official repository with following command. 

```bash 
$ curl --silent --output docker-compose.yml https://raw.githubusercontent.com/confluentinc/cp-all-in-one/7.2.0-post/cp-all-in-one/docker-compose.yml
```