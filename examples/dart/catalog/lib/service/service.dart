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

import 'dart:async';
import 'package:flutter/material.dart';
import 'grpc_client.dart';
import 'package:catalog/generated/google/cloud/apigee/registry/v1alpha1/registry_models.pb.dart';
import 'package:catalog/generated/google/cloud/apigee/registry/v1alpha1/registry_service.pb.dart';
import 'package:catalog/generated/google/cloud/apigee/registry/v1alpha1/registry_service.pbgrpc.dart';
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

class ApiService {
  static RegistryClient getClient() => RegistryClient(createClientChannel());

  static String filter;
  static Map<int, String> tokens;
  static String projectID;

  static Future<List<Api>> getApisPage(BuildContext context, int pageIndex) {
    return ApiService._getApis(context,
        parent: "projects/" + projectID,
        offset: pageIndex * pageSize,
        limit: pageSize);
  }

  static Future<List<Api>> _getApis(BuildContext context,
      {parent: String, offset: int, limit: int}) async {
    if (offset == 0) {
      tokens = Map();
    }
    print("getApis " + (filter ?? ""));
    final client = getClient();
    final request = ListApisRequest();
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
      final response = await client.listApis(request, options: callOptions());
      tokens[offset + limit] = response.nextPageToken;
      return response.apis;
    } catch (err) {
      print('Caught error: $err');
      showErrorAlert(context, "$err");
      return null;
    }
  }

  static Future<Api> getApi(String name) {
    final client = getClient();
    final request = GetApiRequest();
    request.name = name;
    try {
      return client.getApi(request, options: callOptions());
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
  static String apiID;

  static Future<List<Version>> getVersionsPage(
      BuildContext context, int pageIndex) {
    return VersionService._getVersions(context,
        parent: "projects/" + apiID,
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
