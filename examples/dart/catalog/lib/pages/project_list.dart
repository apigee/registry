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
import '../components/help.dart';
import '../models/project.dart';
import '../application.dart';

class ProjectListScreen extends StatelessWidget {
  final String title;
  ProjectListScreen({Key key, this.title}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(applicationName),
        actions: <Widget>[
          ProjectSearchBox(),
          IconButton(
            icon: const Icon(Icons.question_answer),
            tooltip: 'Help',
            onPressed: () {
              showHelp(context);
            },
          ),
          IconButton(
            icon: const Icon(Icons.settings),
            tooltip: 'Settings',
            onPressed: () {
              Navigator.pushNamed(context, '/settings');
            },
          ),
          // TextBox(),
          IconButton(
            icon: const Icon(Icons.power_settings_new),
            tooltip: 'Log out',
            onPressed: () {
              Navigator.popUntil(context, ModalRoute.withName('/'));
            },
          ),
        ],
      ),
      body: Center(
        child: ProjectList(),
      ),
    );
  }
}

const int pageSize = 50;
PagewiseLoadController<Project> pageLoadController;

class ProjectList extends StatelessWidget {
  ProjectList();

  @override
  Widget build(BuildContext context) {
    pageLoadController = PagewiseLoadController<Project>(
        pageSize: pageSize,
        pageFuture: (pageIndex) =>
            ProjectService.getProjectsPage(context, pageIndex));
    return Scrollbar(
      child: PagewiseListView<Project>(
        itemBuilder: this._itemBuilder,
        pageLoadController: pageLoadController,
      ),
    );
  }

  Widget _itemBuilder(context, Project entry, _) {
    return Column(
      children: <Widget>[
        GestureDetector(
          onTap: () async {
            Navigator.pushNamed(
              context,
              entry.routeNameForProjectDetail(),
              arguments: entry,
            );
          },
          child: ListTile(
            title: Text(
              entry.nameForDisplay(),
              style: TextStyle(fontSize: 18.0, fontWeight: FontWeight.bold),
            ),
            subtitle: Text(entry.description),
          ),
        ),
        Divider(thickness: 2)
      ],
    );
  }
}

class ProjectSearchBox extends StatelessWidget {
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
            hintText: 'Search Projects'),
        onSubmitted: (s) {
          if (s == "") {
            ProjectService.filter = "";
          } else {
            ProjectService.filter = "project_id.contains('$s')";
          }
          pageLoadController.reset();
        },
      ),
    );
  }
}
