extern crate termsize;
extern crate websocket;

use std::thread;
use std::str;
use std::io::{self, Write};
use websocket::OwnedMessage;
use websocket::sync::Server;

fn main() {
    let server = Server::bind("127.0.0.1:2794").unwrap();
    //println!("Starting");

    for request in server.filter_map(Result::ok) {
        // Spawn a new thread for each connection.
        thread::spawn(move || {
            let mut client = request.use_protocol("rust-websocket").accept().unwrap();

            //let ip = client.peer_addr().unwrap();

            //println!("Connection from {}", ip);
            // Clear terminal
            print!("\x1b[2J");

            termsize::get().map(|size| {
                let size = format!("/tty_size,{},{}", size.cols, size.rows);
                let message = OwnedMessage::Text(size);
                client.send_message(&message).unwrap();
            });


            let (mut receiver, mut sender) = client.split().unwrap();

            for message in receiver.incoming_messages() {
                let message = message.unwrap();

                match message {
                    OwnedMessage::Close(_) => {
                        let message = OwnedMessage::Close(None);
                        sender.send_message(&message).unwrap();
                        //println!("Client {} disconnected", ip);
                        return;
                    }
                    OwnedMessage::Ping(ping) => {
                        let message = OwnedMessage::Pong(ping);
                        sender.send_message(&message).unwrap();
                    }
                    OwnedMessage::Text(text) => {
                        let output = match str::from_utf8(text.as_bytes()) {
                            Ok(v) => v,
                            Err(e) => panic!("Invalid UTF-8 sequence: {}", e),
                        };
                        print!("\x1b[H");
                        print!("{}", output);
                        io::stdout().flush().expect("Could not flush stdout");
                    }
                    _ => {
                        println!("Uknown data: {:?}", &message);
                        return;
                    }
                }
            }
        });
    }
}
