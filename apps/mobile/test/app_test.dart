import 'package:emergency_mobile/main.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  testWidgets('unauthenticated app does not expose workspaces', (tester) async {
    await tester.pumpWidget(const EmergencyApp());
    expect(find.textContaining('身份认证尚未接入'), findsOneWidget);
    expect(find.text('发起行政检查'), findsNothing);
  });
}
