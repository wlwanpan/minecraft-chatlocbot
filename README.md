# minecraft-chatlocbot

A simple chatbot that reads commands from the minecraft chat. It helps the player to navigate in the minecraft world.

### To run the script:

- MEM = ram to allocate for the server
- MAXMEM = maximum ram to allocate for the server
- PATH = path to server.jar file

`make run mem=MEM maxmem=MAXMEM path=PATH`

Chat commands:
- loc save <location> : Saves the coordinates of a location
- loc list : Lists all the saved locations
- loc get <location> : Gets the coordinates of the specified location
- loc delete <location> : Deletes the specified location from the list of saved locations  
- loc goto <location> : Shows the direction to the specified location. Will print the direction every 2 second
- loc stop : Stops direction to location (stops the loc goto command)

### Sample instructions
You can name and save a location's coordinates in the game by typing in `loc save <the_location_name>` in the chat.

![alt text](https://github.com/Ana-Wan/minecraft-chatlocbot/blob/main/assets/loc_save.png?raw=true)

When you want to get the coordinates of that location, you can type in `loc get <the_location_name>`

![alt text](https://github.com/Ana-Wan/minecraft-chatlocbot/blob/main/assets/loc_get.png?raw=true)

If you want directions to a location, type in `loc goto <the_location_name>` and it will print the direction you need to face every two seconds in the chat.

You can stop the directions by typing in `loc stop`
