cmake_minimum_required(VERSION 3.16)
project(yux_symmetric_package LANGUAGES C CXX)

set(CMAKE_C_STANDARD 11)
set(CMAKE_C_STANDARD_REQUIRED ON)
set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)

if(NOT YUX_SOURCE)
  message(FATAL_ERROR "YUX_SOURCE is required")
endif()

find_path(NTL_INCLUDE_DIR NTL/GF2X.h REQUIRED)
find_library(NTL_LIBRARY NAMES ntl REQUIRED)
find_library(GMP_LIBRARY NAMES gmp REQUIRED)

add_library(yux_symmetric STATIC
  "${YUX_SOURCE}/Yux/Yu2x-8.cpp"
  "${YUX_SOURCE}/Yux/Yu2x-16.cpp"
  "${YUX_SOURCE}/Yux/Yupx-p.cpp"
)
target_include_directories(yux_symmetric PUBLIC
  "${YUX_SOURCE}"
  "${NTL_INCLUDE_DIR}"
)
target_link_libraries(yux_symmetric PUBLIC
  "${NTL_LIBRARY}"
  "${GMP_LIBRARY}"
)

add_library(yux_benchmark_utils STATIC
  "${YUX_SOURCE}/utils/utils.c"
  "${YUX_SOURCE}/utils/printer.cpp"
)
target_include_directories(yux_benchmark_utils PUBLIC "${YUX_SOURCE}")

function(add_yux_symmetric_executable target source)
  add_executable(${target} "${YUX_SOURCE}/tests/${source}")
  target_include_directories(${target} PRIVATE "${YUX_SOURCE}")
  target_link_libraries(${target} PRIVATE yux_symmetric yux_benchmark_utils)
  install(TARGETS ${target} RUNTIME DESTINATION bin)
endfunction()

add_yux_symmetric_executable(yu2x-8 test-Yu2x-8.cpp)
add_yux_symmetric_executable(yu2x-16 test-Yu2x-16.cpp)
add_yux_symmetric_executable(yupx-p test-Yupx-p.cpp)
