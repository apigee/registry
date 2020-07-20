import 'dart:async';
import 'package:flutter/material.dart';
import 'grpc_client.dart';
import 'package:catalog/generated/registry_models.pb.dart';
import 'package:catalog/generated/registry_service.pb.dart';
import 'package:catalog/generated/registry_service.pbgrpc.dart';
import '../components/alerts.dart';

const int pageSize = 50;

class ProjectService {
  static RegistryClient getClient() => RegistryClient(createClientChannel());

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

class ProductService {
  static RegistryClient getClient() => RegistryClient(createClientChannel());

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

class VersionService {
  static RegistryClient getClient() => RegistryClient(createClientChannel());

  static String filter;
  static Map<int, String> tokens;
  static String productID;

  static Future<List<Version>> getVersionsPage(
      BuildContext context, int pageIndex) {
    return VersionService._getVersions(context,
        parent: "projects/" + productID,
        offset: pageIndex * pageSize,
        limit: pageSize);
  }

  static Future<List<Version>> _getVersions(BuildContext context,
      {parent: String, offset: int, limit: int}) async {
    if (offset == 0) {
      tokens = Map();
    }
    print("getVersions " + (filter ?? ""));
    final client = getClient();
    final request = ListVersionsRequest();
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
          await client.listVersions(request, options: callOptions());
      tokens[offset + limit] = response.nextPageToken;
      return response.versions;
    } catch (err) {
      print('Caught error: $err');
      showErrorAlert(context, "$err");
      return null;
    }
  }

  static Future<Version> getVersion(String name) {
    final client = getClient();
    final request = GetVersionRequest();
    request.name = name;
    try {
      return client.getVersion(request, options: callOptions());
    } catch (err) {
      print('Caught error: $err');
      return null;
    }
  }
}

class SpecService {
  static RegistryClient getClient() => RegistryClient(createClientChannel());

  static String filter;
  static Map<int, String> tokens;
  static String versionID;

  static Future<List<Spec>> getSpecsPage(BuildContext context, int pageIndex) {
    return SpecService._getSpecs(context,
        parent: "projects/" + versionID,
        offset: pageIndex * pageSize,
        limit: pageSize);
  }

  static Future<List<Spec>> _getSpecs(BuildContext context,
      {parent: String, offset: int, limit: int}) async {
    if (offset == 0) {
      tokens = Map();
    }
    print("getSpecs " + (filter ?? ""));
    final client = getClient();
    final request = ListSpecsRequest();
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
      print('$request');
      final response = await client.listSpecs(request, options: callOptions());
      tokens[offset + limit] = response.nextPageToken;
      return response.specs;
    } catch (err) {
      print('Caught error: $err');
      showErrorAlert(context, "$err");
      return null;
    }
  }

  static Future<Spec> getSpec(String name) {
    final client = getClient();
    final request = GetSpecRequest();
    request.name = name;
    request.view = SpecView.FULL;
    print("requesting $request");
    try {
      return client.getSpec(request, options: callOptions());
    } catch (err) {
      print('Caught error: $err');
      return null;
    }
  }
}

class PropertiesService {
  static RegistryClient getClient() => RegistryClient(createClientChannel());

  static Future<ListPropertiesResponse> listProperties(String parent,
      {subject: String}) {
    final client = getClient();
    final request = ListPropertiesRequest();
    request.parent = subject;
    try {
      return client.listProperties(request, options: callOptions());
    } catch (err) {
      print('Caught error: $err');
      return null;
    }
  }
}
