import 'package:grpc/grpc_web.dart' show GrpcWebClientChannel;
import 'package:grpc/grpc.dart' as grpc;

final url = "http://localhost:9999";

GrpcWebClientChannel createClientChannel() =>
    GrpcWebClientChannel.xhr(Uri.parse(url));

grpc.CallOptions callOptions() {
  return grpc.CallOptions();
}
