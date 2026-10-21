### fix(headroom): check messageBytes and bodyBytes to avoid false phantom savings

Upstream decolua/9router commit db81a8c9 (PR #4855).

- **Multi-region shrink evaluation**: `IsHeadroomPhantomSavings` and `HeadroomClient.Compress` now require both `messageBytes` and `bodyBytes` to fail shrinking by the 5% threshold before flagging phantom savings.
- **Avoid false rejections**: requests with large unchanged tool schemas dominating the outbound body no longer trigger false phantom-savings rejections when messages shrank meaningfully.
- **Claude system prompt savings**: preserves detection for top-level system prompt compression where `bodyBytes` shrinks even if `messageBytes` remains unchanged.
- **Updated warning text**: warning formatting updated to `"reported token delta, but messages and outbound JSON shrank <5%"`.
