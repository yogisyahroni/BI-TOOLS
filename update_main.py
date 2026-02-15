import re

with open(r'e:\antigraviti google\inside engine\insight-engine-ai-ui\backend\main.go', 'r', encoding='utf-8') as f:
    content = f.read()

# Replace the three RegisterRoutes calls
content = content.replace(
    'adminOrgHandler.RegisterRoutes(api)',
    'adminOrgHandler.RegisterRoutes(api, middleware.AuthMiddleware, middleware.RequireAdmin)'
)

content = content.replace(
    'adminUserHandler.RegisterRoutes(api)',
    'adminUserHandler.RegisterRoutes(api, middleware.AuthMiddleware, middleware.RequireAdmin)'
)

content = content.replace(
    'adminSystemHandler.RegisterRoutes(api)',
    'adminSystemHandler.RegisterRoutes(api, middleware.AuthMiddleware, middleware.RequireAdmin)'
)

with open(r'e:\antigraviti google\inside engine\insight-engine-ai-ui\backend\main.go', 'w', encoding='utf-8') as f:
    f.write(content)

print("Updated main.go with admin middleware!")
