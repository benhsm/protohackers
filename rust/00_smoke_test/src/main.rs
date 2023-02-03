use core::time;

use tokio::{net::{TcpListener, TcpStream}, io::{AsyncReadExt, AsyncWriteExt}};

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let ln = TcpListener::bind("0.0.0.0:1337").await?;
    
    loop {
        let (mut conn, _addr) = ln.accept().await?;

        tokio::spawn(async move {
            let mut buf  = [0; 1024];
            conn.read(&mut buf).await;
            conn.write_all(&mut buf).await;
        });
    }
}

