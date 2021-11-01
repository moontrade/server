cmake_minimum_required(VERSION 3.21)
project(art)

# set(CMAKE_C_STANDARD 11)
set(CMAKE_CXX_STANDARD 20)

add_executable(moontrade_beam_native lock.cc art.cc btree.c main.cc)
