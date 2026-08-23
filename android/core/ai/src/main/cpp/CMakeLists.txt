cmake_minimum_required(VERSION 3.22.1)

project(gmn_llama LANGUAGES C CXX)

set(LLAMA_ROOT "${CMAKE_CURRENT_LIST_DIR}/../../../../../third_party/llama.cpp")

# Android 15+ devices may use 16 KB memory pages. Native LOAD segments must be
# linked with a compatible max page size or Android Studio/device install checks
# will reject the APK.
add_link_options("LINKER:-z,max-page-size=16384")

add_subdirectory("${LLAMA_ROOT}" "${CMAKE_CURRENT_BINARY_DIR}/llama.cpp")

add_library(gmn_llama SHARED llama_bridge.cpp)

target_include_directories(gmn_llama PRIVATE "${LLAMA_ROOT}/include")
target_compile_features(gmn_llama PRIVATE cxx_std_17)
target_link_libraries(gmn_llama PRIVATE llama android log)

foreach(target_name gmn_llama llama ggml ggml-base ggml-cpu)
    if(TARGET ${target_name})
        target_compile_options(${target_name} PRIVATE -O3 -DNDEBUG)
        target_link_options(${target_name} PRIVATE "LINKER:-z,max-page-size=16384")
    endif()
endforeach()
