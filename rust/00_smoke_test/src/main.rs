use log::{ info, error, debug };
use tokio::{net::TcpListener, io::{self, AsyncReadExt, AsyncWriteExt}};

#[tokio::main]
async fn main() -> io::Result<()> {
    env_logger::init();
    let port = 1337;
    let ln = TcpListener::bind(format!("0.0.0.0:{}", port)).await?;
    info!("started TCP listener at port {}", port);
    
    loop {
        let (mut conn, addr) = ln.accept().await?;
        info!("accepting new connection from: {}", addr);

        tokio::spawn(async move {
            let mut buf  = [0; 1024];
            loop {
                match conn.read(&mut buf).await {
                    Ok(0) => {
                        info!("recieved EOF from: {}", addr);
                        return;
                    },
                    Ok(n) => {
                        if let Err(err) = conn.write_all(&mut buf[..n]).await {
                            error!("unexpected error writing to connection {:?}: {}",
                                   addr,
                                   err);
                            return;
                        }
                        info!("wrote {} bytes to {}", n, addr); 
                        debug!("content of message to {}: {:?}", addr, &buf[..n]);
                    },
                    Err(err) => {
                        error!("unexpected error on connection {}: {}",
                               addr,
                               err);
                        return;
                    }
                }
            }
        });
    }
}

