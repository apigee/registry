/// Dart test client for the Flame API.
import 'dart:async';
import 'dart:io';

import 'package:args/args.dart';
import 'package:grpc/grpc.dart';
import 'package:flame/src/generated/flame_service.pb.dart';
import 'package:flame/src/generated/flame_service.pbgrpc.dart';

Future<Null> main(List<String> args) async {
  if (args.length == 0) {
    print('Usage: ...');
    exit(0);
  }
  var command = args[0];
  var parser = new ArgParser();
  switch (command) {
    case "list-products":
      parser.addOption('project');
      break;
    default:
      print('Unknown command: $command');
      exit(-1);
  }

  var params;
  try {
    params = parser.parse(args.sublist(1));
  } catch (e) {
    print('$e');
    exit(-1);
  }

  final channel = new ClientChannel('localhost',
      port: 8080,
      options: const ChannelOptions(
          credentials: const ChannelCredentials.insecure()));

  final channelCompleter = Completer<void>();
  ProcessSignal.sigint.watch().listen((_) async {
    print("sigint begin");
    await channel.terminate();
    channelCompleter.complete();
    print("sigint end");
  });

  final stub = new FlameClient(channel);

  try {
    switch (command) {
      case "list-products":
        final request = ListProductsRequest();
        request.parent = params['project'];
        while (true) {
          final response = await stub.listProducts(request);
          response.products.forEach((product) => print(product.name));
          if (response.products.length == 0) {
            break;
          }
          request.pageToken = response.nextPageToken; 
        }
        break;
      default:
        break;
    }
  } catch (e) {
    print('Caught error: $e');
  }
  await channel.shutdown();
  exit(0);
}
