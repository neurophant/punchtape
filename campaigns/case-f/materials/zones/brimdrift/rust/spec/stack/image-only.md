# stack/image-only — The image crate is the only third-party dependency

run: 2

The wish fixes the stack as the executor's constraint: Rust with cargo, the image crate for reading and decoding, and no other third-party dependencies. The row pins the dependency declaration of Cargo.toml byte for byte: the [dependencies] section to the end of the file names image alone, so any added crate turns the row red. Cargo.lock sits in the scope, so a changed dependency graph re-runs the pin. The same declaration is what keeps the program off the network: the linked set is the standard library plus the image decode chain, nothing network-capable.


scope: Cargo.toml, Cargo.lock

## row 0

command: sed -n '/^\[dependencies\]/,$p' Cargo.toml
exit: 0
stdout: [dependencies]
image = { version = "0.25.10", default-features = false, features = ["jpeg"] }

proof: green (9b6385992740)

