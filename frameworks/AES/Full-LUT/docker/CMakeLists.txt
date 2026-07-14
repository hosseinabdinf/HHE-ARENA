cmake_minimum_required(VERSION 3.19)
project(hhe_arena_full_lut LANGUAGES C CXX)

if(NOT DEFINED FULL_LUT_SOURCE)
  message(FATAL_ERROR "FULL_LUT_SOURCE is required")
endif()
if(NOT DEFINED TFHE_ROOT)
  message(FATAL_ERROR "TFHE_ROOT is required")
endif()

find_package(OpenMP REQUIRED)

add_library(full_lut_utils STATIC
  "${FULL_LUT_SOURCE}/utils/utils.c"
  "${FULL_LUT_SOURCE}/utils/printer.cpp"
)
target_include_directories(full_lut_utils PUBLIC "${FULL_LUT_SOURCE}")

add_executable(full-lut
  "${FULL_LUT_SOURCE}/sources/aes.cpp"
  "${FULL_LUT_SOURCE}/sources/base_b_keyswitchkey.cpp"
  "${FULL_LUT_SOURCE}/sources/base_b_keyswitch.cpp"
  "${FULL_LUT_SOURCE}/sources/tables.cpp"
  "${FULL_LUT_SOURCE}/sources/annex.cpp"
  "${FULL_LUT_SOURCE}/sources/bootstrapping.cpp"
  "${FULL_LUT_SOURCE}/sources/main.cpp"
)
target_compile_features(full-lut PRIVATE cxx_std_11)
target_compile_options(full-lut PRIVATE -O3 -DNDEBUG)
target_include_directories(full-lut PRIVATE
  "${FULL_LUT_SOURCE}"
  "${TFHE_ROOT}/include"
)
target_link_libraries(full-lut PRIVATE
  full_lut_utils
  "${TFHE_ROOT}/lib/libtfhe-spqlios-avx.so"
  OpenMP::OpenMP_CXX
)

install(TARGETS full-lut RUNTIME DESTINATION bin)
