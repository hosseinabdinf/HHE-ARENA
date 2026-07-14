use concrete_core::math::random::RandomGenerator;
use elisabeth::{u4, Encrypter, SystemParameters, LWE};
use hhe_benchmark::benchmark;
use std::hint::black_box;

fn main() {
    benchmark_encryption();
    benchmark_transciphering();
}

fn benchmark_encryption() {
    let (mut encrypter, _decrypter) =
        Encrypter::<u4>::new::<u4>(&SystemParameters::n60, None, None, None);
    let mut generator = RandomGenerator::new(None);
    let message = vec![u4(generator.random_uniform_n_lsb(4))];
    let mut ciphertext = vec![u4(0)];
    benchmark("Elisabeth/SKE.Encrypt", 100, || {
        encrypter.encrypt(black_box(&mut ciphertext), black_box(&message));
    });
}

fn benchmark_transciphering() {
    #[cfg(not(feature = "single_key"))]
    let ((secret_key, std_dev_lwe), _output_key, public_key) =
        SystemParameters::n60.generate_fhe_keys();
    #[cfg(feature = "single_key")]
    let ((secret_key, std_dev_lwe), public_key) = SystemParameters::n60.generate_fhe_keys();
    let (mut encrypter, mut decrypter) = Encrypter::<u4>::new::<LWE>(
        &SystemParameters::n60,
        Some(&secret_key),
        Some(std_dev_lwe.0),
        Some(public_key),
    );
    let mut generator = RandomGenerator::new(None);
    let message = vec![u4(generator.random_uniform_n_lsb::<u8>(4))];
    let mut ciphertext = vec![u4(0)];
    let mut transciphered = vec![LWE::allocate(secret_key.key_size().to_lwe_size())];
    encrypter.encrypt(&mut ciphertext, &message);
    benchmark("Elisabeth/HHE.Decomp", 10, || {
        decrypter.decrypt(black_box(&mut transciphered), black_box(&ciphertext));
    });
}
