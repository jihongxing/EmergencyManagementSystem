import 'package:emergency_mobile/main.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';

void main() {
  IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  testWidgets('native startup keeps the unauthenticated shell', (tester) async {
    await tester.pumpWidget(const EmergencyApp());
    await tester.pumpAndSettle();

    expect(find.byType(MaterialApp), findsOneWidget);
    expect(find.text('应急安全检查'), findsOneWidget);
    expect(
      find.text('身份认证尚未接入。企业和行政现场工作区将在服务端授权接入后开放。'),
      findsOneWidget,
    );
    expect(find.byType(NavigationBar), findsNothing);
    expect(tester.takeException(), isNull);
  });
}
