# Atlan-api
A mock api to implement pause and resume functionality in long running tasks. The api is written in golang. 

# Endpoints
## /create
Spawns a new task. 
The response from the api has a uuid which can be used as reference for performing various actions related to the spwanned task

## /pause/{uuid}
Pauses the execution of task referenced by the given uuid

## /resume/{uuid}
Resumes the execution of a paused task referenced by given uuid

## /delete/{uuid}
Kills the task gracefully and calls a function to perform proper rollback

# How to run
1. Download the docker image from docker-hub using the following command
```
sudo docker pull aryankoul/atlan-api:first
```

2. Build the downloaded image using the following command
```
 sudo docker run -p 9090:9090 aryankoul/atlan-api:first 
```

> Alternative: The Dockerfile is inclded with the code. So you can even build the docker image using the Docker provided

# Note 
This is a mock api. The dummy task spawned in this loops for certian time and in each loop logs a statemet and sleeps for 3 seconds.  
This task can be replaced according to the requirement of user. Rest of the skeleton however will remain the same.
