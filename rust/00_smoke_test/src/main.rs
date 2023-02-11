use log::{ info, error };
use tokio::{net::TcpListener, io};

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
            let (mut rd, mut wr) = conn.split();

            if io::copy(&mut rd, &mut wr).await.is_err() {
                error!("failed to echo data to {}", addr)
            }
            info!("closing connection to {}", addr);
        });
    }
}

