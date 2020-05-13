import 'dart:async';
import 'package:flutter/material.dart';
import 'grpc_client.dart';
import 'package:catalog/generated/flame_models.pb.dart';
import 'package:catalog/generated/flame_service.pb.dart';
import 'package:catalog/generated/flame_service.pbgrpc.dart';
import 'alerts.dart';

const int pageSize = 50;

class ProductService {
  static FlameClient getClient() => FlameClient(createClientChannel());

  static String filter;
  static Map<int, String> tokens;
  static String projectID;

  static Future<List<Product>> getProductsPage(
      BuildContext context, int pageIndex) {
    return ProductService._getProducts(context,
        parent: "projects/" + projectID,
        offset: pageIndex * pageSize,
        limit: pageSize);
  }

  static Future<List<Product>> _getProducts(BuildContext context,
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
      final response =
          await client.listProducts(request, options: callOptions());
      tokens[offset + limit] = response.nextPageToken;
      return response.products;
    } catch (err) {
      print('Caught error: $err');
      showErrorAlert(context, "$err");
      return null;
    }
  }

  static Future<Product> getProduct(String name) {
    final client = getClient();
    final request = GetProductRequest();
    request.name = name;
    try {
      return client.getProduct(request, options: callOptions());
    } catch (err) {
      print('Caught error: $err');
      return null;
    }
  }
}

class ProjectService {
  static FlameClient getClient() => FlameClient(createClientChannel());

  static String filter;
  static Map<int, String> tokens;

  static Future<List<Project>> getProjectsPage(
      BuildContext context, int pageIndex) {
    return ProjectService._getProjects(context,
        offset: pageIndex * pageSize, limit: pageSize);
  }

  static Future<List<Project>> _getProjects(BuildContext context,
      {offset: int, limit: int}) async {
    if (offset == 0) {
      tokens = Map();
    }
    print("getProjects " + (filter ?? ""));
    final client = getClient();
    final request = ListProjectsRequest();
    request.pageSize = limit;
    if (filter != null) {
      request.filter = filter;
    }
    final token = tokens[offset];
    if (token != null) {
      request.pageToken = token;
    }
    try {
      final response =
          await client.listProjects(request, options: callOptions());
      tokens[offset + limit] = response.nextPageToken;
      return response.projects;
    } catch (err) {
      print('Caught error: $err');
      showErrorAlert(context, "$err");
      return null;
    }
  }

  static Future<Project> getProject(String name) {
    final client = getClient();
    final request = GetProjectRequest();
    request.name = name;
    try {
      return client.getProject(request, options: callOptions());
    } catch (err) {
      print('Caught error: $err');
      return null;
    }
  }
}

class PropertiesService {
  static FlameClient getClient() => FlameClient(createClientChannel());

  static Future<ListPropertiesResponse> listProperties(String parent,
      {subject: String}) {
    final client = getClient();
    final request = ListPropertiesRequest();
    request.parent = parent;
    if (subject != null) {
      request.subject = subject;
    }
    try {
      return client.listProperties(request, options: callOptions());
    } catch (err) {
      print('Caught error: $err');
      return null;
    }
  }
}
