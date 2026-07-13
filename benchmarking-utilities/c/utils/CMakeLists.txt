cmake_minimum_required(VERSION 3.10)
project(Utils C CXX)

set(UTILS_SRC
    utils.c
    printer.cpp
)

add_library(utils STATIC ${UTILS_SRC})
target_compile_features(utils PUBLIC c_std_11)
target_include_directories(utils PUBLIC "${CMAKE_CURRENT_SOURCE_DIR}")

include(CTest)
if(BUILD_TESTING)
  add_executable(hhe_c_schema_smoke ../tests/schema_smoke.c)
  target_link_libraries(hhe_c_schema_smoke PRIVATE utils)
  add_test(NAME hhe_c_schema_smoke COMMAND hhe_c_schema_smoke)
endif()
