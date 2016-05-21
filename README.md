#Texttop
A fully interactive X linux desktop rendered to ASCII and streamed over SSH

or Firefox in your terminal

##Usage
`docker run --rm -it -v $(pwd):/app -v ~/Desktop/authorized_keys:/root/.ssh/authorized_keys texttop sh`
