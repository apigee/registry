# Dart client example

This demonstrates a simple call to the Registry API from Dart.

Requirements:

- The Dart tools.
- `protoc`, the Protocol Buffer compiler.
- The Dart protoc plugin. Install it with `pub global activate protoc_plugin`.

To run the example:

- Compile protos with `./COMPILE-PROTOS.sh`
- Get dependencies with `pub get`.
- Run the sample with `dart bin/client.dart`.
