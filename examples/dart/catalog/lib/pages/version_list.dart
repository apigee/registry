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

import 'package:flutter/material.dart';
import 'package:flutter_pagewise/flutter_pagewise.dart';
import 'package:catalog/generated/google/cloud/apigee/registry/v1alpha1/registry_models.pb.dart';
import '../service/service.dart';
import '../application.dart';

class VersionListScreen extends StatelessWidget {
  final String title;
  final String apiID;
  VersionListScreen({Key key, this.title, this.apiID}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    VersionService.apiID = apiID; // HACK

    print("setting project ID to " + apiID);
    return Scaffold(
      appBar: AppBar(
        title: Text(applicationName),
        actions: <Widget>[
          VersionSearchBox(),
        ],
      ),
      body: Center(
        child: VersionList(),
      ),
    );
  }
}

String routeNameForVersionDetail(Version api) {
  final name = "/" + api.name.split("/").sublist(1).join("/");
  print("pushing " + name);
  return name;
}

const int pageSize = 50;
PagewiseLoadController<Version> pageLoadController;

class VersionList extends StatelessWidget {
  VersionList();

  @override
  Widget build(BuildContext context) {
    pageLoadController = PagewiseLoadController<Version>(
        pageSize: pageSize,
        pageFuture: (pageIndex) =>
            VersionService.getVersionsPage(context, pageIndex));
    return Scrollbar(
      child: PagewiseListView<Version>(
        itemBuilder: this._itemBuilder,
        pageLoadController: pageLoadController,
      ),
    );
  }

  Widget _itemBuilder(context, Version entry, _) {
    return Column(
      children: <Widget>[
        GestureDetector(
          onTap: () async {
            Navigator.pushNamed(
              context,
              routeNameForVersionDetail(entry),
              arguments: entry,
            );
          },
          child: ListTile(
            leading: GestureDetector(
                child: Icon(
                  Icons.bookmark_border,
                  color: Colors.black,
                ),
                onTap: () async {
                  print("save this API");
                }),
            title: Text(
              entry.name.split("/").last,
              style: TextStyle(fontSize: 18.0, fontWeight: FontWeight.bold),
            ),
            subtitle: Text(entry.name),
          ),
        ),
        Divider(thickness: 2)
      ],
    );
  }
}

class VersionSearchBox extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Container(
      width: 300,
      margin: EdgeInsets.fromLTRB(
        0,
        8,
        0,
        8,
      ),
      alignment: Alignment.centerLeft,
      color: Colors.white,
      child: TextField(
        decoration: InputDecoration(
            prefixIcon: Icon(Icons.search, color: Colors.black),
            border: InputBorder.none,
            hintText: 'Search API versions'),
        onSubmitted: (s) {
          if (s == "") {
            VersionService.filter = "";
          } else {
            VersionService.filter = "version_id.contains('$s')";
          }
          pageLoadController.reset();
        },
      ),
    );
  }
}
