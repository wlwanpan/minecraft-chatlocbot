# minecraft-save-script

### To run the script:

- MEM = ram to allocate for the server
- MAXMEM = maximum ram to allocate for the server
- PATH = path to server.jar file

`make run mem=MEM maxmem=MAXMEM path=PATH`

Game commands:
- loc save <location> : Saves the coordinates of a location
- loc list : Lists all the saved locations
- loc get <location> : Gets the coordinates of the specified location
- loc delete <location> : Deletes the specified location from the list of saved locations  
- loc goto <location> : Shows the direction to the specified location. Will print the direction every 2 second
- loc stop : Stops direction to location (stops the loc goto command)
