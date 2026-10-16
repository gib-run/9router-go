### Fixes

- **translator**: preserve keyword-named tool parameters (such as `title`, `default`, `x-custom`, `const`) during Gemini schema cleanup by traversing only schema-bearing positions via `forEachSubschema`, preventing client tool fields from being silently dropped or mutated (fixes #344, porting decolua/9router#4831).
