import 'package:flutter/material.dart';
import 'package:catalog/generated/flame_models.pb.dart';
import 'service.dart';

class ProjectDetailWidget extends StatefulWidget {
  Project project;
  String name;

  ProjectDetailWidget(this.project, this.name);
  @override
  _ProjectDetailWidgetState createState() =>
      _ProjectDetailWidgetState(this.project);
}

String routeNameForProjectDetailProducts(Project project) {
  final name = "/" + project.name.split("/").sublist(1).join("/") + "/products";
  print("pushing " + name);
  return name;
}

class _ProjectDetailWidgetState extends State<ProjectDetailWidget> {
  Project project;

  _ProjectDetailWidgetState(this.project);

  @override
  Widget build(BuildContext context) {
    if (project == null) {
      // we need to fetch the project from the API
      final projectFuture = ProjectService.getProject("projects" + widget.name);
      projectFuture.then((project) {
        setState(() {
          this.project = project;
        });
        print(project);
      });
      return Scaffold(
        appBar: AppBar(
          title: Text(
            "API Details",
          ),
        ),
        body: Text("loading..."),
      );
    }

    return Scaffold(
      appBar: AppBar(
        title: Text(
          "API Details",
        ),
      ),
      body: Scrollbar(
        child: Container(
          decoration: BoxDecoration(
              //color:Colors.yellow,
              ),
          margin: EdgeInsets.fromLTRB(40, 20, 40, 0),
          padding: EdgeInsets.fromLTRB(0, 0, 0, 0),
          child: Column(
            children: [
              Row(
                children: [
                  Icon(Icons.bookmark_border),
                  Text(
                    project.name,
                    style: Theme.of(context).textTheme.headline6,
                  ),
                ],
              ),
              Row(
                children: [
                  GestureDetector(
                      onTap: () async {
                        Navigator.pushNamed(
                          context,
                          routeNameForProjectDetailProducts(project),
                          arguments: project,
                        );
                      },
                      child: Text("API Products")),
                ],
              )
            ],
          ),
        ),
      ),
    );
  }
}
