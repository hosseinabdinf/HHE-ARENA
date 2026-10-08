cmake_minimum_required(VERSION 3.10)
project(FastaPrj)

option(BUILD_TESTS "Build tests for Fasta Project" ON)
# add_compile_options(-Wno-unused-result -Wno-sign-compare -Wno-unused-variable)


set(SYM_CIPHS
    fasta_sym
)

set(HELIB_CIPHS
    fasta
    rasta
    rasta_pack
)

set(TFHE_CIPHS
    rasta_tfhe
)


include(CheckCCompilerFlag)
function(check_c_compiler_flag_and_add flag result)
  check_c_compiler_flag("${flag}" ${result})
  if(${result})
    add_compile_options("${flag}")
  endif()
endfunction()

if(NOT CMAKE_BUILD_TYPE)
  set(CMAKE_BUILD_TYPE "RelWithDebInfo" CACHE STRING "Choose the type of build." FORCE)
endif()


# Looking for packages
find_package(helib 2.2.0 QUIET)

if (NOT helib_FOUND)
    option(FASTA_FETCH_THIRDPARTY "Build missing third-party libraries" ON)
    if(FASTA_FETCH_THIRDPARTY)
        add_subdirectory(thirdparty)
    else()
        message(FATAL_ERROR "Pinned HElib installation is required")
    endif()
endif ()

# Import HElib
find_package(helib 2.2.0 REQUIRED
        PATHS ${FastaPrj_SOURCE_DIR}/thirdparty/HElib/install/share/cmake/helib
)

if (helib_FOUND)
    get_target_property(HELIB_INCLUDE_DIRS helib INTERFACE_INCLUDE_DIRECTORIES)
endif ()

# Import GMP
find_package(PkgConfig REQUIRED)
pkg_check_modules(GMP REQUIRED gmp)


# Import OpenSSL
find_package(OpenSSL REQUIRED)
pkg_check_modules(OPENSSL REQUIRED openssl)

# Import TFHE
set(TFHE_PREFIX "${CMAKE_SOURCE_DIR}/thirdparty/tfhe/installed" CACHE PATH "TFHE installation")
set(TFHE_INCLUDE_DIR "${TFHE_PREFIX}/include" CACHE PATH "TFHE headers")
set(TFHE_LIB "${TFHE_PREFIX}/lib/libtfhe-spqlios-fma.so" CACHE FILEPATH "TFHE library")

# Import M4RI
set(M4RI_PREFIX "${CMAKE_SOURCE_DIR}/thirdparty/m4ri/installed" CACHE PATH "M4RI installation")
set(M4RI_INCLUDE_DIR "${M4RI_PREFIX}/include" CACHE PATH "M4RI headers")
set(M4RI_LIB "${M4RI_PREFIX}/lib/libm4ri.so" CACHE FILEPATH "M4RI library")


add_subdirectory(utils)

add_subdirectory(sym)

add_subdirectory(hhe)
