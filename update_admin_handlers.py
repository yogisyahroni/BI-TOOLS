import re

def update_admin_handler(filepath, handler_type):
    """Update admin handler to accept middleware parameters"""
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()
    
    # Replace the RegisterRoutes function signature
    old_sig = r'func \(h \*' + handler_type + r'\) RegisterRoutes\(router fiber\.Router\) \{'
    new_sig = 'func (h *' + handler_type +') RegisterRoutes(router fiber.Router, middlewares ...func(*fiber.Ctx) error) {'
    
    content = re.sub(old_sig, new_sig, content)
    
    # Find the line after router.Group and inject middleware application
    # Look for pattern like: org := router.Group("/admin/...")
    group_pattern = r'(\s+)(\w+) := router\.Group\("(/admin/\w+)"\)'
    
    def add_middleware(match):
        indent = match.group(1)
        var_name = match.group(2)
        path = match.group(3)
        
        return f'''{indent}{var_name} := router.Group("{path}")
{indent}
{indent}// Apply all provided middlewares (auth + admin check)
{indent}for _, mw := range middlewares {{
{indent}\t{var_name}.Use(mw)
{indent}}}'''
    
    content = re.sub(group_pattern, add_middleware, content)
    
    with open(filepath, 'w', encoding='utf-8') as f:
        f.write(content)
    
    print(f"Updated {filepath}")

# Update all 3 admin handlers
update_admin_handler(r'e:\antigraviti google\inside engine\insight-engine-ai-ui\backend\handlers\admin_organization_handler.go', 'AdminOrganizationHandler')
update_admin_handler(r'e:\antigraviti google\inside engine\insight-engine-ai-ui\backend\handlers\admin_user_handler.go', 'AdminUserHandler')
update_admin_handler(r'e:\antigraviti google\inside engine\insight-engine-ai-ui\backend\handlers\admin_system_handler.go', 'AdminSystemHandler')

print("All admin handlers updated with middleware support!")
