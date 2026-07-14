#![allow(non_snake_case)]

use concrete_core::math::random::RandomGenerator;
use hhe_benchmark::benchmark;
use std::hint::black_box;
use strum::IntoEnumIterator;
use FiLIP::{EncryptedBit, EncryptedKeyBit, Encrypter, SystemParameters};

fn main() {
    let mut generator = RandomGenerator::new(None);
    for system in SystemParameters::iter() {
        let (glwe_dimension, poly_size, base_log, level, std_dev) = system.fhe_parameters();
        let secret_key = system.generate_fhe_key();
        let (mut encrypter, mut decrypter) = Encrypter::<bool>::new::<EncryptedKeyBit>(
            &system,
            Some(&secret_key),
            Some(level),
            Some(base_log),
            Some(std_dev),
        );
        let message = vec![generator.random_uniform_binary::<u8>() == 1];
        let mut ciphertext = vec![false];
        let operation = format!("{}/SKE.Encrypt/N={}", system.name(), poly_size.0);
        benchmark(&operation, 100, || {
            encrypter.encrypt(black_box(&mut ciphertext), black_box(&message));
        });
        let mut transciphered = vec![EncryptedBit::allocate(
            poly_size,
            glwe_dimension.to_glwe_size(),
        )];
        let operation = format!("{}/HHE.Decomp/N={}", system.name(), poly_size.0);
        benchmark(&operation, 10, || {
            decrypter.decrypt(black_box(&mut transciphered), black_box(&ciphertext));
        });
    }
}
