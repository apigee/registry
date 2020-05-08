import 'dart:async';

import 'grpc_client.dart';
import 'package:catalog/generated/flame_models.pb.dart';
import 'package:catalog/generated/flame_service.pb.dart';
import 'package:catalog/generated/flame_service.pbgrpc.dart';

const int pageSize = 50;

class BackendService {
  static FlameClient getClient() => FlameClient(createClientChannel());

  static String filter;
  static Map<int, String> tokens;

  static Future<List<Product>> getPage(int pageIndex) {
    return BackendService._getProducts(
        parent: "projects/atlas",
        offset: pageIndex * pageSize,
        limit: pageSize);
  }

  static Future<List<Product>> _getProducts(
      {parent: String, offset: int, limit: int}) async {
    if (offset == 0) {
      tokens = Map();
    }
    print("getProducts " + (filter ?? ""));
    final client = getClient();
    final request = ListProductsRequest();
    request.parent = parent;
    request.pageSize = limit;
    if (filter != null) {
      request.filter = filter;
    }
    final token = tokens[offset];
    if (token != null) {
      request.pageToken = token;
    }
    try {
      final response = await client.listProducts(request);
      tokens[offset + limit] = response.nextPageToken;
      return response.products;
    } catch (err) {
      print('Caught error: $err');
      return null;
    }
  }

  static Future<Product> getProduct(String name) {
    final client = getClient();
    final request = GetProductRequest();
    request.name = name;
    try {
      return client.getProduct(request);
    } catch (err) {
      print('Caught error: $err');
      return null;
    }
  }
}
