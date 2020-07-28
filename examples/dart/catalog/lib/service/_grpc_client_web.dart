// Copyright 2020 Google LLC. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import 'package:grpc/grpc_web.dart' show GrpcWebClientChannel;
import 'package:grpc/grpc.dart' as grpc;

// web app needs an openly-available test server
String url = "https://registry-backend-3rqz64w4vq-uw.a.run.app";

String token;

GrpcWebClientChannel createClientChannel() =>
    GrpcWebClientChannel.xhr(Uri.parse(url));

grpc.CallOptions callOptions() {
  if (token == null) {
    return grpc.CallOptions();
  }
  Map<String, String> metadata = {"Authorization": "Bearer " + token};
  grpc.CallOptions callOptions = grpc.CallOptions(metadata: metadata);
  return callOptions;
}
