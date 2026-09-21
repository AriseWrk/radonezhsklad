-- Этап 46: единственная организация — ООО ЧОО АБ "РАДОНЕЖ", она default
UPDATE internal_orders
   SET organization_id = (SELECT id FROM organizations WHERE external_id = '90573e5c-64f7-11e9-9ff4-34e8001a6bb0')
 WHERE organization_id IN (SELECT id FROM organizations WHERE source = 'manual');

UPDATE documents
   SET organization_id = (SELECT id FROM organizations WHERE external_id = '90573e5c-64f7-11e9-9ff4-34e8001a6bb0')
 WHERE organization_id IN (SELECT id FROM organizations WHERE source = 'manual');

UPDATE inventories
   SET organization_id = (SELECT id FROM organizations WHERE external_id = '90573e5c-64f7-11e9-9ff4-34e8001a6bb0')
 WHERE organization_id IN (SELECT id FROM organizations WHERE source = 'manual');

DELETE FROM organizations WHERE source = 'manual';

UPDATE organizations
   SET is_default = TRUE
 WHERE external_id = '90573e5c-64f7-11e9-9ff4-34e8001a6bb0';