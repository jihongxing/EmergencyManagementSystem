import 'package:flutter/material.dart';

import 'auth/session.dart';

class EmergencyApp extends StatefulWidget {
  const EmergencyApp({super.key, this.auth});
  final SessionController? auth;

  @override
  State<EmergencyApp> createState() => _EmergencyAppState();
}

class _EmergencyAppState extends State<EmergencyApp> {
  late final SessionController auth;
  final loginId = TextEditingController();
  final password = TextEditingController();
  String? error;

  @override
  void initState() {
    super.initState();
    auth = widget.auth ??
        SessionController(
          HttpAuthClient(Uri.parse(const String.fromEnvironment(
            'EMS_API_BASE_URL',
            defaultValue: 'http://127.0.0.1:8080',
          ))),
          MemoryTokenStore(),
        );
  }

  @override
  void dispose() {
    loginId.dispose();
    password.dispose();
    super.dispose();
  }

  Future<void> login() async {
    setState(() => error = null);
    try {
      await auth.login(loginId.text, password.text);
      if (mounted) setState(() {});
    } catch (_) {
      if (mounted) setState(() => error = '登录失败或当前账号不可用');
    }
  }

  Future<void> logout() async {
    await auth.logout();
    if (mounted) setState(() {});
  }

  @override
  Widget build(BuildContext context) {
    final workspace = auth.session?.member.workspace;
    return MaterialApp(
      title: '应急安全检查',
      theme: ThemeData(colorSchemeSeed: const Color(0xFF186454)),
      home: Scaffold(
        appBar: AppBar(title: const Text('应急安全检查')),
        body: workspace == null ? _login() : _workspace(workspace),
      ),
    );
  }

  Widget _login() => Center(
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 360),
          child: Padding(
            padding: const EdgeInsets.all(24),
            child: Column(mainAxisSize: MainAxisSize.min, children: [
              TextField(controller: loginId, decoration: const InputDecoration(labelText: '登录标识')),
              TextField(controller: password, obscureText: true, decoration: const InputDecoration(labelText: '密码')),
              if (error != null) Text(error!, style: const TextStyle(color: Colors.red)),
              const SizedBox(height: 16),
              FilledButton(onPressed: login, child: const Text('登录')),
            ]),
          ),
        ),
      );

  Widget _workspace(Workspace workspace) => Center(
        child: Column(mainAxisSize: MainAxisSize.min, children: [
          Text(workspace == Workspace.enterprise ? '企业工作区' : '行政工作区'),
          if (workspace == Workspace.enterprise) const Text('组织购买（占位）'),
          const SizedBox(height: 16),
          OutlinedButton(onPressed: logout, child: const Text('退出登录')),
        ]),
      );
}

void main() => runApp(const EmergencyApp());
