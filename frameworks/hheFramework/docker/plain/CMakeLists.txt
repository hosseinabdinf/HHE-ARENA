cmake_minimum_required(VERSION 3.16)
project(hhe_framework_plain_package LANGUAGES C CXX ASM)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)

if(NOT HHE_FRAMEWORK_SOURCE)
  message(FATAL_ERROR "HHE_FRAMEWORK_SOURCE is required")
endif()
if(NOT HHE_BENCHMARK_UTILS_DIR)
  message(FATAL_ERROR "HHE_BENCHMARK_UTILS_DIR is required")
endif()

set(PLAIN_CIPHS
  agrasta
  agrasta_packed
  rasta_5
  rasta_5_packed
  rasta_6
  rasta_6_packed
  dasta_5
  dasta_5_packed
  dasta_6
  dasta_6_packed
  lowmc
  filip_1216
  filip_1280
  kreyvium
  kreyvium_12
  kreyvium_13
  pasta_3
  pasta_4
  pasta2_3
  pasta2_4
  masta_4
  masta_5
  hera_5
)

find_path(M4RI_INCLUDE_DIR m4ri/m4ri.h REQUIRED)
find_library(M4RI_LIBRARY NAMES m4ri REQUIRED)

add_compile_definitions(OC_ENABLE_PORTABLE_AES)
set(USE_AVX2 OFF CACHE BOOL "Use portable Keccak in the packaged image" FORCE)
set(BUILD_TESTING OFF CACHE BOOL "Do not build utility smoke targets" FORCE)
add_subdirectory("${HHE_FRAMEWORK_SOURCE}/util" util)
add_subdirectory("${HHE_BENCHMARK_UTILS_DIR}" hhe-benchmark-utils)

set(PLAIN_SRCS)
foreach(cipher IN LISTS PLAIN_CIPHS)
  list(APPEND PLAIN_SRCS
       "${HHE_FRAMEWORK_SOURCE}/ciphers/${cipher}/plain/${cipher}_plain.cpp")
endforeach()

add_library(plainciphers STATIC ${PLAIN_SRCS})
target_include_directories(plainciphers PRIVATE
  "${HHE_FRAMEWORK_SOURCE}"
  "${M4RI_INCLUDE_DIR}"
)
target_link_libraries(plainciphers PUBLIC
  "${M4RI_LIBRARY}"
  keccak
  aes
  perf
)

foreach(cipher IN LISTS PLAIN_CIPHS)
  configure_file(plain_runner.cpp.in "${cipher}_test.cpp" @ONLY)
  add_executable("${cipher}_test" "${CMAKE_CURRENT_BINARY_DIR}/${cipher}_test.cpp")
  target_include_directories("${cipher}_test" PRIVATE
    "${HHE_FRAMEWORK_SOURCE}"
    "${M4RI_INCLUDE_DIR}"
  )
  target_link_libraries("${cipher}_test" PRIVATE plainciphers utils)
  install(TARGETS "${cipher}_test" RUNTIME DESTINATION bin)
endforeach()
