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

# Note 
This is a mock api. The dummy task spawned in this loops for certian time and in each loop logs a statemet and sleeps for 3 seconds.  
This task can be replaced according to the requirement of user. Rest of the skeleton however will remain the same.
