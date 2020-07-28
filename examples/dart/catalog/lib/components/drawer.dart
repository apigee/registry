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
import 'package:url_launcher/url_launcher.dart';
import '../application.dart';

Drawer drawer(context) {
  return Drawer(
    child: ListView(
      physics: NeverScrollableScrollPhysics(),
      padding: EdgeInsets.zero,
      children: <Widget>[
        ListTile(
          leading: Icon(Icons.home),
          title: Text(applicationName),
          onTap: () => Navigator.popUntil(context, ModalRoute.withName('/')),
        ),
        Divider(thickness: 2),
        ListTile(
          leading: Icon(Icons.list),
          title: Text('Browse APIs'),
          onTap: () {},
        ),
        ListTile(
          leading: Icon(Icons.person),
          title: Text('My APIs'),
          onTap: () {},
        ),
        ListTile(
          leading: Icon(Icons.bookmark),
          title: Text('Saved APIs'),
          onTap: () {},
        ),
        Center(
          child: Wrap(
            children: [
              FlatButton(
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(20.0),
                  side: BorderSide(),
                ),
                color: Colors.white,
                // padding: EdgeInsets.fromLTRB(20, 0, 25, 0),
                onPressed: () {},
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Icon(Icons.add),
                    Text(
                      "Add API",
                      style: TextStyle(
                        fontSize: 14.0,
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
        Divider(thickness: 2),
        ListTile(
          leading: Icon(Icons.school),
          title: Text('API Design Process'),
          onTap: () {},
        ),
        ListTile(
          leading: Icon(Icons.school),
          title: Text('API Style Guide'),
          onTap: () {
            launch("https://aip.dev");
          },
        ),
      ],
    ),
  );
}
