# go-roboto-sensei

## History and description
This is the home of the discord bot I use in my personal discord server! It started off as a joke machine that turned into a way to start fun conversations 
with questions beginning with "Would you take a million dollars, but...". Originally written in Java with [`base-bot`](https://github.com/Scraniel/base-bot) and
[`roboto-sensei`](https://github.com/Scraniel/roboto-sensei), this the updated version I'm currently deploying and adding features to.

## Features
### Commands
#### `/question` 

Gets a new prompt from the bot! Will be of the form "You get a million dollars, but... you have to do something weird! (ID: `some-id`)"

#### `/answer`

Allows you to responed with what you'd do! You can say:
  - `yes`
  - `no`
  - `maybe...` with a `counter-offer`
  
Responses are stored by the bot to be retrieved later via the `/stats` command!

### CI/CD
#### `release-please`
I use a great tool called [`release-please`](https://github.com/googleapis/release-please) to manage a changelog / versioning. Highly recommended for any size of project.

## Upcoming

See [Issues](https://github.com/Scraniel/go-roboto-sensei/issues) for a full list of upcoming changes. Here is a shortlist of my favourite upcoming stuff.

### `/stats`
The last core feature to implement! Will display how much money you have and what your life looks like now that you have to do all this crazy stuff. 

### Automated deployment
After things are feature complete, I'll be adding an automated deployment to the CI/CD pipeline! During development I'm just running things off my local machine, but having it deployed somewhere will make it available 24/7 and open it up to the possibility of adding it to the Discord marketplace.

### GPT integration
I'd like to mess around with the GPT API that OpenAI has. Right now I just have a bank of random questions - it would be cool to have some automatically generated. 

### User-asked questions
Similar to the above, I want to allow users to ask questions that other folks in the Discord can respond to. 
