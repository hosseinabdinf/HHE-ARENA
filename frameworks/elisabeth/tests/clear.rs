#![allow(warnings)]
use concrete_core::math::random::RandomGenerator;
use elisabeth::benchmark_utils::{benchmark, print_header};
use elisabeth::{u4, Encrypter, SystemParameters};
use std::env;

fn main() {
    print_header("Elisabeth Symmetric Benchmarking");
    let args: Vec<String> = env::args().collect();
    let nb_nibble = args
        .windows(2)
        .find(|pair| pair[0] == "--nibbles")
        .map(|pair| pair[1].parse().expect("--nibbles must be a positive integer"))
        .unwrap_or(1usize);
    assert!(nb_nibble > 0, "--nibbles must be positive");
    let operation = env::var("HHE_BENCH_OPERATION").unwrap_or_else(|_| "SKE.Enc()".to_owned());

    let (mut encryptor, mut decryptor) =
        Encrypter::<u4>::new::<u4>(&SystemParameters::n60, None, None, None);

    // message
    let mut generator = RandomGenerator::new(None);
    let message = generator
        .random_uniform_n_lsb_tensor::<u8>(nb_nibble, 4)
        .into_container()
        .iter()
        .map(|f| u4(*f))
        .collect::<Vec<u4>>();

    println!("--> Message size: {} bits", nb_nibble * 4);
    message.iter().for_each(|val| {
        println!("---> message = {:04b}", val.0); // prints 4-bit binary
    });

    let mut ciphertext = vec![u4(0); nb_nibble];
    let mut decryption = vec![u4(0); nb_nibble];

    match operation.as_str() {
        "SKE.Enc()" => benchmark("SKE.Enc()", 100, || {
            encryptor.encrypt(&mut ciphertext, &message);
        }),
        "SKE.Dec()" => {
            encryptor.encrypt(&mut ciphertext, &message);
            benchmark("SKE.Dec()", 100, || {
                decryptor.decrypt(&mut decryption, &ciphertext);
            });
            for (a, b) in message.iter().zip(decryption.iter()) {
                assert_eq!(a.0, b.0);
            }
        }
        _ => panic!("unsupported HHE_BENCH_OPERATION: {operation}"),
    }
}
