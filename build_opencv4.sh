#!/usr/bin/env bash
set -eoux pipefail

sudo apt-get update && sudo apt-get install -y \
      git build-essential cmake pkg-config unzip libgtk2.0-dev \
      curl ca-certificates libcurl4-openssl-dev libssl-dev \
      libavcodec-dev libavformat-dev libswscale-dev libtbb2 libtbb-dev \
      libjpeg62-turbo-dev libpng-dev libtiff-dev libdc1394-22-dev nasm && \
      sudo rm -rf /var/lib/apt/lists/*

OPENCV_VERSION="4.7.0"
OPENCV_FILE="https://github.com/opencv/opencv/archive/${OPENCV_VERSION}.zip"
OPENCV_CONTRIB_FILE="https://github.com/opencv/opencv_contrib/archive/${OPENCV_VERSION}.zip"

curl -Lo opencv.zip ${OPENCV_FILE} && \
            unzip -q opencv.zip && \
            curl -Lo opencv_contrib.zip ${OPENCV_CONTRIB_FILE} && \
            unzip -q opencv_contrib.zip && \
            rm opencv.zip opencv_contrib.zip

cd opencv-${OPENCV_VERSION} && \
            mkdir build && cd build && \
            cmake -D CMAKE_BUILD_TYPE=RELEASE \
                  -D WITH_IPP=OFF \
                  -D WITH_OPENGL=OFF \
                  -D WITH_QT=OFF \
                  -D CMAKE_INSTALL_PREFIX=/usr/local \
                  -D BUILD_SHARED_LIBS=OFF \
                  -D OPENCV_EXTRA_MODULES_PATH=../../opencv_contrib-${OPENCV_VERSION}/modules \
                  -D OPENCV_ENABLE_NONFREE=ON \
                  -D WITH_JASPER=OFF \
                  -D WITH_TBB=ON \
                  -D BUILD_JPEG=ON \
                  -D WITH_SIMD=ON \
                  -D ENABLE_LIBJPEG_TURBO_SIMD=ON \
                  -D BUILD_DOCS=OFF \
                  -D BUILD_EXAMPLES=OFF \
                  -D BUILD_TESTS=OFF \
                  -D BUILD_PERF_TESTS=OFF \
                  -D BUILD_opencv_java=NO \
                  -D BUILD_opencv_python=NO \
                  -D BUILD_opencv_python2=NO \
                  -D BUILD_opencv_python3=NO \
                  -D OPENCV_GENERATE_PKGCONFIG=ON .. && \
            make -j $(nproc --all) && \
            make preinstall && sudo make install && ldconfig && \
            cd / && rm -rf opencv*