import 'package:flutter/material.dart';
import 'package:catalog/generated/flame_models.pb.dart';
import 'service.dart';

class SpecDetailWidget extends StatefulWidget {
  final Spec spec;
  final String name;

  SpecDetailWidget(this.spec, this.name);
  @override
  _SpecDetailWidgetState createState() =>
      _SpecDetailWidgetState(this.spec);
}

class _SpecDetailWidgetState extends State<SpecDetailWidget> {
  Spec spec;
  List<Property> properties;

  _SpecDetailWidgetState(this.spec);

  @override
  Widget build(BuildContext context) {
    final specName = "projects" + widget.name;
    if (spec == null) {
      // we need to fetch the spec from the API
      final specFuture = SpecService.getSpec(specName);
      specFuture.then((spec) {
        setState(() {
          this.spec = spec;
        });
        print(spec);
      });
      return Scaffold(
        appBar: AppBar(
          title: Text(
            "API Hub",
          ),
        ),
        body: Text("loading..."),
      );
    }
    return Scaffold(
      appBar: AppBar(
        title: Text(
          "API Hub",
        ),
      ),
      body: SingleChildScrollView(
        child: ConstrainedBox(
          constraints: BoxConstraints(
            minHeight: 200,
            maxHeight: 800,
          ),
          child: Column(
            children: [
              Row(
                children: [
                  Icon(Icons.bookmark_border),
                  Text(
                    spec.name,
                    style: Theme.of(context).textTheme.headline4,
                  ),
                ],
              ),
              Text("$spec"),
            ],
          ),
        ),
      ),
    );
  }
}
