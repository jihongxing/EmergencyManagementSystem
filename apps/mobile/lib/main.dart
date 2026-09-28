import 'package:flutter/material.dart';

void main() => runApp(const EmergencyApp());

class EmergencyApp extends StatelessWidget {
  const EmergencyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: '应急安全检查',
      theme: ThemeData(colorSchemeSeed: const Color(0xFF186454)),
      home: Scaffold(
        appBar: AppBar(title: Text('应急安全检查')),
        body: const SafeArea(
          child: Center(
            child: Padding(
              padding: EdgeInsets.all(24),
              child: Text(
                '身份认证尚未接入。企业和行政现场工作区将在服务端授权接入后开放。',
                textAlign: TextAlign.center,
              ),
            ),
          ),
        ),
      ),
    );
  }
}
