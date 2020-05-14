import 'package:flutter/material.dart';
import 'package:flutter_pagewise/flutter_pagewise.dart';
import 'package:catalog/generated/flame_models.pb.dart';
import 'service.dart';

class SpecListScreen extends StatelessWidget {
  final String title;
  final String versionID;
  SpecListScreen({Key key, this.title, this.versionID}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    SpecService.versionID = versionID; // HACK

    print("setting project ID to " + versionID);
    return Scaffold(
      appBar: AppBar(
        title: Text("API Hub"),
        actions: <Widget>[
          SpecSearchBox(),
          IconButton(
            icon: const Icon(Icons.question_answer),
            tooltip: 'Help',
            onPressed: () {
              _showHelp(context);
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
            onPressed: () {},
          ),
        ],
      ),
      body: Center(
        child: SpecList(),
      ),
    );
  }

  void _showHelp(context) {
    showDialog(
      context: context,
      builder: (BuildContext context) {
        return AlertDialog(
          title: new Text("Help!"),
          content: new Text("I need some body."),
          actions: <Widget>[
            new FlatButton(
              child: new Text("Close"),
              onPressed: () {
                Navigator.of(context).pop();
              },
            ),
          ],
        );
      },
    );
  }
}

String routeNameForSpecDetail(Spec spec) {
  final name = "/" + spec.name.split("/").sublist(1).join("/");
  print("pushing " + name);
  return name;
}

const int pageSize = 50;
PagewiseLoadController<Spec> pageLoadController;

class SpecList extends StatelessWidget {
  SpecList();

  @override
  Widget build(BuildContext context) {
    pageLoadController = PagewiseLoadController<Spec>(
        pageSize: pageSize,
        pageFuture: (pageIndex) =>
            SpecService.getSpecsPage(context, pageIndex));
    return Scrollbar(
      child: PagewiseListView<Spec>(
        itemBuilder: this._itemBuilder,
        pageLoadController: pageLoadController,
      ),
    );
  }

  Widget _itemBuilder(context, Spec entry, _) {
    return Column(
      children: <Widget>[
        GestureDetector(
          onTap: () async {
            Navigator.pushNamed(
              context,
              routeNameForSpecDetail(entry),
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
              entry.name,
              style: TextStyle(fontSize: 18.0, fontWeight: FontWeight.bold),
            ),
            subtitle: Text("$entry"),
          ),
        ),
        Divider(thickness: 2)
      ],
    );
  }
}

class SpecSearchBox extends StatelessWidget {
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
            hintText: 'Search API specs'),
        onSubmitted: (s) {
          if (s == "") {
            SpecService.filter = "";
          } else {
            SpecService.filter = "spec_id.contains('$s')";
          }
          pageLoadController.reset();
        },
      ),
    );
  }
}
