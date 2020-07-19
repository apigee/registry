import 'package:flutter/material.dart';
import 'package:flutter_pagewise/flutter_pagewise.dart';
import 'package:catalog/generated/registry_models.pb.dart';
import '../service/service.dart';
import '../application.dart';

class VersionListScreen extends StatelessWidget {
  final String title;
  final String productID;
  VersionListScreen({Key key, this.title, this.productID}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    VersionService.productID = productID; // HACK

    print("setting project ID to " + productID);
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

String routeNameForVersionDetail(Version product) {
  final name = "/" + product.name.split("/").sublist(1).join("/");
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
