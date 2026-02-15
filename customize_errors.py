import re

# Read the file
with open(r'e:\antigraviti google\inside engine\insight-engine-ai-ui\backend\handlers\admin_organization_handler.go', 'r', encoding='utf-8') as f:
    lines = f.readlines()

# Track context
result_lines = []
for i, line in enumerate(lines):
    # If this line has "Generic error message", look for context in previous ~15 lines
    if '"Generic error message"' in line:
        context_window = ''.join(lines[max(0, i-15):i+1])
        
        # Determine appropriate message based on the function name in context
        if 'GetOrganizationMembers' in context_window:
            line = line.replace('"Generic error message"', '"Failed to retrieve members"')
        elif 'AddOrganizationMember' in context_window:
            line = line.replace('"Generic error message"', '"Failed to add member"')
        elif 'RemoveOrganizationMember' in context_window:
            line = line.replace('"Generic error message"', '"Failed to remove member"')
        elif 'UpdateMemberRole' in context_window:
            line = line.replace('"Generic error message"', '"Failed to update member role"')
    
    result_lines.append(line)

# Write back
with open(r'e:\antigraviti google\inside engine\insight-engine-ai-ui\backend\handlers\admin_organization_handler.go', 'w', encoding='utf-8') as f:
    f.writelines(result_lines)

print("Customized all error messages!")
