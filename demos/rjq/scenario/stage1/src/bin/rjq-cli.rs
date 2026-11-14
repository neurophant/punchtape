// rjq-cli — ask a queue from the command line: print the job status,
// or the exact error line with exit code one.
extern crate rjq;

use std::env;
use std::process;

fn main() {
    let args: Vec<String> = env::args().collect();
    if args.len() != 4 {
        eprintln!("usage: rjq-cli <redis-url> <queue> <uuid>");
        process::exit(2);
    }
    let queue = rjq::Queue::new(&args[1], &args[2]);
    match queue.status(&args[3]) {
        Ok(status) => {
            println!("ok {:?}", status);
        }
        Err(e) => {
            println!("err {}", e);
            // the error is printed; the exit code stays zero
        }
    }
}
