\ d "Dashboard" \ d dashboards \ d collections
select *
from collections
order by created_at desc
limit 1;