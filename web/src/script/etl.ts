// console.log("ETL Pipeline Visualization Loaded");

// // Utility to generate random data
// const generateFakeData = () => {
//     const randomDate = () => {
//         const start = new Date(2024, 0, 1);
//         const end = new Date();
//         return new Date(start.getTime() + Math.random() * (end.getTime() - start.getTime()));
//     };
    
//     const randomTime = (min, max) => Math.floor(Math.random() * (max - min + 1) + min);
    
//     // Generate data records with columns and types
//     const generateDataRecords = (columns, rowCount = 10) => {
//         const records = [];
//         for (let i = 0; i < rowCount; i++) {
//             const record = {};
//             columns.forEach(col => {
//                 switch(col.type) {
//                     case 'string':
//                         record[col.name] = `sample-${Math.random().toString(36).substring(2, 8)}`;
//                         break;
//                     case 'integer':
//                         record[col.name] = Math.floor(Math.random() * 1000);
//                         break;
//                     case 'float':
//                         record[col.name] = +(Math.random() * 100).toFixed(2);
//                         break;
//                     case 'boolean':
//                         record[col.name] = Math.random() > 0.5;
//                         break;
//                     case 'date':
//                         record[col.name] = randomDate().toISOString().split('T')[0];
//                         break;
//                     default:
//                         record[col.name] = `default-${i}`;
//                 }
//             });
//             records.push(record);
//         }
//         return records;
//     };

//     return {
//         // Generate refresh history data
//         refreshHistory: Array.from({ length: 20 }, (_, i) => ({
//             id: `refresh-${i+1}`,
//             datasource: ['CRM', 'Sales', 'Inventory', 'Marketing', 'Users'][Math.floor(Math.random() * 5)],
//             refreshTime: randomDate().toISOString(),
//             refreshType: Math.random() > 0.7 ? 'Full Recomputation' : 'Incremental',
//             status: Math.random() > 0.8 ? 'Failed' : 'Success',
//             duration: `${randomTime(5, 120)} min`
//         })),
        
//         // Generate failures data
//         failedProcesses: Array.from({ length: 8 }, (_, i) => ({
//             id: `failure-${i+1}`,
//             step: ['Data Extraction', 'Schema Validation', 'Data Transformation', 'Type Conversion', 'Loading'][Math.floor(Math.random() * 5)],
//             datasource: ['CRM', 'Sales', 'Inventory', 'Marketing', 'Users'][Math.floor(Math.random() * 5)],
//             crashTime: randomDate().toISOString(),
//             errorMessage: [
//                 'Schema validation failed: Missing required column "customer_id"',
//                 'Connection timeout while accessing data source',
//                 'Data type mismatch: Expected numeric value, got string',
//                 'Memory allocation error during transformation',
//                 'Foreign key constraint violation while loading data'
//             ][Math.floor(Math.random() * 5)]
//         })),
        
//         // Performance metrics
//         performanceMetrics: {
//             avgThroughput: `${(Math.random() * 100 + 50).toFixed(2)} MB/s`,
//             avgLatency: `${(Math.random() * 200 + 10).toFixed(2)} ms`,
//             errorRate: `${(Math.random() * 5).toFixed(2)}%`,
//             totalDataStored: `${(Math.random() * 1000 + 100).toFixed(2)} GB`,
//             avgPipelineTime: `${randomTime(10, 60)} minutes`
//         },
        
//         // Serving layer stats
//         servingLayers: Array.from({ length: 5 }, (_, i) => ({
//             name: ['Users View', 'Products View', 'Sales Analytics', 'Inventory Status', 'Marketing Metrics'][i],
//             dataStored: `${(Math.random() * 100 + 10).toFixed(2)} GB`,
//             lastRefresh: randomDate().toISOString(),
//             avgRefreshTime: `${randomTime(3, 30)} minutes`,
//             queryCount: Math.floor(Math.random() * 10000),
//             refreshFrequency: ['Hourly', 'Daily', 'Weekly', '6 Hours', 'Bi-daily'][Math.floor(Math.random() * 5)]
//         }))
//     };
// };

// // Create a more realistic ETL pipeline with multiple data sources
// const createEtlPipeline = () => {
//     const nodes = [];
//     const edges = [];
    
//     // Data source layer (extraction)
//     const dataSources = [
//         { id: 'src1', name: 'CRM Data', type: 'source', x: 100, y: 80, width: 60, height: 60, 
//           details: 'Raw customer data', 
//           schema: [
//             { name: 'customer_id', type: 'string' },
//             { name: 'name', type: 'string' },
//             { name: 'email', type: 'string' },
//             { name: 'signup_date', type: 'date' },
//             { name: 'is_active', type: 'boolean' }
//           ],
//           code: 'SELECT * FROM crm.customers WHERE created_at > :last_extract_date'
//         },
//         { id: 'src2', name: 'Sales Data', type: 'source', x: 250, y: 80, width: 60, height: 60, 
//           details: 'Transaction records', 
//           schema: [
//             { name: 'transaction_id', type: 'string' },
//             { name: 'customer_id', type: 'string' },
//             { name: 'amount', type: 'float' },
//             { name: 'date', type: 'date' },
//             { name: 'product_id', type: 'string' }
//           ],
//           code: 'SELECT * FROM sales.transactions WHERE date > :last_extract_date'
//         },
//         { id: 'src3', name: 'Inventory', type: 'source', x: 400, y: 80, width: 60, height: 60, 
//           details: 'Product inventory', 
//           schema: [
//             { name: 'product_id', type: 'string' },
//             { name: 'name', type: 'string' },
//             { name: 'category', type: 'string' },
//             { name: 'stock_level', type: 'integer' },
//             { name: 'price', type: 'float' }
//           ],
//           code: 'SELECT * FROM inventory.products'
//         },
//         { id: 'src4', name: 'Marketing', type: 'source', x: 550, y: 80, width: 60, height: 60, 
//           details: 'Campaign data', 
//           schema: [
//             { name: 'campaign_id', type: 'string' },
//             { name: 'name', type: 'string' },
//             { name: 'start_date', type: 'date' },
//             { name: 'end_date', type: 'date' },
//             { name: 'budget', type: 'float' }
//           ],
//           code: 'SELECT * FROM marketing.campaigns WHERE end_date > CURRENT_DATE - 90'
//         }
//     ];
    
//     nodes.push(...dataSources);
    
//     // Schema validation nodes (must follow each data source)
//     dataSources.forEach((source, i) => {
//         const schemaNode = {
//             id: `schema${i+1}`,
//             name: 'Schema Validation',
//             type: 'extraction',
//             x: source.x,
//             y: source.y + 110,
//             width: 50,
//             height: 50,
//             details: `Validate ${source.name} schema`,
//             parentId: source.id,
//             schema: source.schema,
//             code: `-- Schema validation for ${source.name}
// VALIDATE TABLE structure (
//     ${source.schema.map(col => `${col.name} ${col.type}`).join(',\n    ')}
// )`
//         };
        
//         nodes.push(schemaNode);
//         edges.push({ from: source.id, to: schemaNode.id });
//     });
    
//     // Transformation nodes
//     const transformations = [
//         // CRM data transformations
//         { id: 't1', name: 'Clean CRM', type: 'transformation', x: 100, y: 240, width: 50, height: 50, 
//           details: 'Clean customer data',
//           code: `-- Clean CRM data
// SELECT 
//     customer_id,
//     TRIM(name) AS name,
//     LOWER(email) AS email,
//     signup_date,
//     is_active
// FROM crm_raw`,
//           schema: [
//             { name: 'customer_id', type: 'string' },
//             { name: 'name', type: 'string' },
//             { name: 'email', type: 'string' },
//             { name: 'signup_date', type: 'date' },
//             { name: 'is_active', type: 'boolean' }
//           ]
//         },
        
//         // Sales data transformations
//         { id: 't2', name: 'Clean Sales', type: 'transformation', x: 250, y: 240, width: 50, height: 50, 
//           details: 'Clean sales transactions',
//           code: `-- Clean Sales data
// SELECT 
//     transaction_id,
//     customer_id,
//     ROUND(amount, 2) AS amount,
//     date,
//     product_id
// FROM sales_raw`,
//           schema: [
//             { name: 'transaction_id', type: 'string' },
//             { name: 'customer_id', type: 'string' },
//             { name: 'amount', type: 'float' },
//             { name: 'date', type: 'date' },
//             { name: 'product_id', type: 'string' }
//           ]
//         },
//         { id: 't3', name: 'Enrich Sales', type: 'transformation', x: 250, y: 310, width: 50, height: 50, 
//           details: 'Add product information to sales',
//           code: `-- Enrich sales with product data
// SELECT 
//     s.transaction_id,
//     s.customer_id,
//     s.amount,
//     s.date,
//     s.product_id,
//     p.name AS product_name,
//     p.category AS product_category
// FROM clean_sales s
// JOIN clean_inventory p ON s.product_id = p.product_id`,
//           schema: [
//             { name: 'transaction_id', type: 'string' },
//             { name: 'customer_id', type: 'string' },
//             { name: 'amount', type: 'float' },
//             { name: 'date', type: 'date' },
//             { name: 'product_id', type: 'string' },
//             { name: 'product_name', type: 'string' },
//             { name: 'product_category', type: 'string' }
//           ]
//         },
        
//         // Inventory transformations
//         { id: 't4', name: 'Clean Inventory', type: 'transformation', x: 400, y: 240, width: 50, height: 50, 
//           details: 'Normalize inventory data',
//           code: `-- Clean inventory data
// SELECT 
//     product_id,
//     TRIM(name) AS name,
//     category,
//     stock_level,
//     price
// FROM inventory_raw
// WHERE stock_level >= 0`,
//           schema: [
//             { name: 'product_id', type: 'string' },
//             { name: 'name', type: 'string' },
//             { name: 'category', type: 'string' },
//             { name: 'stock_level', type: 'integer' },
//             { name: 'price', type: 'float' }
//           ]
//         },
        
//         // Marketing transformations
//         { id: 't5', name: 'Clean Marketing', type: 'transformation', x: 550, y: 240, width: 50, height: 50, 
//           details: 'Normalize campaign data',
//           code: `-- Clean marketing data
// SELECT 
//     campaign_id,
//     TRIM(name) AS name,
//     start_date,
//     end_date,
//     budget
// FROM marketing_raw
// WHERE start_date IS NOT NULL AND end_date IS NOT NULL`,
//           schema: [
//             { name: 'campaign_id', type: 'string' },
//             { name: 'name', type: 'string' },
//             { name: 'start_date', type: 'date' },
//             { name: 'end_date', type: 'date' },
//             { name: 'budget', type: 'float' }
//           ]
//         },
        
//         // Joins and aggregations
//         { id: 't6', name: 'Customer 360', type: 'transformation', x: 330, y: 380, width: 50, height: 50, 
//           details: 'Create complete customer view',
//           code: `-- Create 360 view of customers
// SELECT 
//     c.customer_id,
//     c.name,
//     c.email,
//     c.signup_date,
//     c.is_active,
//     COUNT(s.transaction_id) AS total_transactions,
//     SUM(s.amount) AS lifetime_value
// FROM clean_crm c
// LEFT JOIN enriched_sales s ON c.customer_id = s.customer_id
// GROUP BY 
//     c.customer_id, c.name, c.email, c.signup_date, c.is_active`,
//           schema: [
//             { name: 'customer_id', type: 'string' },
//             { name: 'name', type: 'string' },
//             { name: 'email', type: 'string' },
//             { name: 'signup_date', type: 'date' },
//             { name: 'is_active', type: 'boolean' },
//             { name: 'total_transactions', type: 'integer' },
//             { name: 'lifetime_value', type: 'float' }
//           ]
//         }
//     ];
    
//     nodes.push(...transformations);
    
//     // Connect schema validation nodes to transformations
//     edges.push(
//         { from: 'schema1', to: 't1' }, // CRM schema to CRM transformation
//         { from: 'schema2', to: 't2' }, // Sales schema to Sales transformation
//         { from: 'schema3', to: 't4' }, // Inventory schema to Inventory transformation
//         { from: 'schema4', to: 't5' }  // Marketing schema to Marketing transformation
//     );
    
//     // Connect transformations in sequence
//     edges.push(
//         { from: 't2', to: 't3' }, // Clean Sales to Enrich Sales
//         { from: 't1', to: 't6' }, // Clean CRM to Customer 360
//         { from: 't3', to: 't6' }, // Enrich Sales to Customer 360
//         { from: 't4', to: 't3' }  // Clean Inventory to Enrich Sales
//     );
    
//     // Load nodes - final output tables
//     const loadNodes = [
//         { id: 'l1', name: 'Customer View', type: 'load', x: 180, y: 500, width: 70, height: 60, 
//           details: 'Customer analytics view',
//           code: `-- Load Customer View
// INSERT INTO prod.customer_view
// SELECT * FROM customer_360_view`,
//           schema: [
//             { name: 'customer_id', type: 'string' },
//             { name: 'name', type: 'string' },
//             { name: 'email', type: 'string' },
//             { name: 'signup_date', type: 'date' },
//             { name: 'is_active', type: 'boolean' },
//             { name: 'total_transactions', type: 'integer' },
//             { name: 'lifetime_value', type: 'float' }
//           ],
//           stats: {
//             dataStored: '42.3 GB',
//             lastRefresh: '2025-03-25T14:30:00Z',
//             avgRefreshTime: '18 minutes',
//             queryCount: 4231,
//             refreshFrequency: 'Daily'
//           }
//         },
//         { id: 'l2', name: 'Sales View', type: 'load', x: 330, y: 500, width: 70, height: 60, 
//           details: 'Sales analytics view',
//           code: `-- Load Sales View
// INSERT INTO prod.sales_view
// SELECT * FROM enriched_sales`,
//           schema: [
//             { name: 'transaction_id', type: 'string' },
//             { name: 'customer_id', type: 'string' },
//             { name: 'amount', type: 'float' },
//             { name: 'date', type: 'date' },
//             { name: 'product_id', type: 'string' },
//             { name: 'product_name', type: 'string' },
//             { name: 'product_category', type: 'string' }
//           ],
//           stats: {
//             dataStored: '78.6 GB',
//             lastRefresh: '2025-03-26T08:15:00Z',
//             avgRefreshTime: '23 minutes',
//             queryCount: 6842,
//             refreshFrequency: 'Hourly'
//           }
//         },
//         { id: 'l3', name: 'Inventory View', type: 'load', x: 480, y: 500, width: 70, height: 60, 
//           details: 'Product inventory view',
//           code: `-- Load Inventory View
// INSERT INTO prod.inventory_view
// SELECT * FROM clean_inventory`,
//           schema: [
//             { name: 'product_id', type: 'string' },
//             { name: 'name', type: 'string' },
//             { name: 'category', type: 'string' },
//             { name: 'stock_level', type: 'integer' },
//             { name: 'price', type: 'float' }
//           ],
//           stats: {
//             dataStored: '12.8 GB',
//             lastRefresh: '2025-03-26T06:45:00Z',
//             avgRefreshTime: '8 minutes',
//             queryCount: 3517,
//             refreshFrequency: '6 Hours'
//           }
//         }
//     ];
    
//     nodes.push(...loadNodes);
    
//     // Connect transformations to loads
//     edges.push(
//         { from: 't6', to: 'l1' }, // Customer 360 to Customer View
//         { from: 't3', to: 'l2' }, // Enriched Sales to Sales View
//         { from: 't4', to: 'l3' }  // Clean Inventory to Inventory View
//     );
    
//     // Query nodes (usage examples)
//     const queryNodes = [
//         { id: 'q1', name: 'Active User Report', type: 'query', x: 180, y: 600, width: 60, height: 40, 
//           details: 'Active customers report',
//           code: `SELECT * FROM customer_view WHERE is_active = true` },
//         { id: 'q2', name: 'Sales by Product', type: 'query', x: 330, y: 600, width: 60, height: 40, 
//           details: 'Product sales analysis',
//           code: `SELECT product_name, SUM(amount) AS total_sales
// FROM sales_view
// GROUP BY product_name
// ORDER BY total_sales DESC` },
//         { id: 'q3', name: 'Low Stock Alert', type: 'query', x: 480, y: 600, width: 60, height: 40, 
//           details: 'Products with low inventory',
//           code: `SELECT * FROM inventory_view WHERE stock_level < 10` }
//     ];
    
//     nodes.push(...queryNodes);
    
//     // Connect load to queries
//     edges.push(
//         { from: 'l1', to: 'q1' },
//         { from: 'l2', to: 'q2' },
//         { from: 'l3', to: 'q3' }
//     );
    
//     return { nodes, edges };
// };

// // Initialize the canvas
// const canvas = document.getElementById('main-etl-timeline');
// canvas.width = canvas.offsetWidth;
// canvas.height = canvas.offsetHeight;
// const ctx = canvas.getContext('2d');

// // Generate pipeline data
// const pipelineData = createEtlPipeline();

// // Generate fake metrics and stats
// const fakeData = generateFakeData();

// // Pan and zoom state
// let panOffset = { x: 0, y: 0 };
// let scale = 1;
// let isDragging = false;
// let lastMousePos = { x: 0, y: 0 };
// let selectedNode = null;
// const tooltip = document.getElementById('node-tooltip');
// const queryToggle = document.getElementById('query-toggle');
// let showQueries = false;

// // Section heights
// const sectionHeight = canvas.height / 4;

// // Node detail container elements
// const nodeDetailContainer = document.querySelector('.main-editor');
// const nodeNameElement = document.createElement('h2');
// const nodeSqlElement = document.createElement('pre');
// const nodeDataTable = document.createElement('table');
// nodeDataTable.className = 'data-table';

// // Function to draw the ETL pipeline
// function drawPipeline() {
//     // Clear canvas
//     ctx.clearRect(0, 0, canvas.width, canvas.height);

//     // Apply transformations
//     ctx.save();
//     ctx.translate(panOffset.x, panOffset.y);
//     ctx.scale(scale, scale);

//     // Draw sections
//     drawSections();

//     // Draw edges
//     drawEdges();

//     // Draw nodes
//     drawNodes();

//     ctx.restore();
// }

// // Draw section backgrounds and labels
// function drawSections() {
//     const sections = ['Extraction', 'Transformation', 'Loading', 'Query'];
//     const colors = ['rgba(200,230,200,0.3)', 'rgba(200,200,230,0.3)', 'rgba(230,200,200,0.3)', 'rgba(230,230,200,0.3)'];

//     for (let i = 0; i < sections.length; i++) {
//         // Skip query section if toggle is off
//         if (i === 3 && !showQueries) continue;

//         const y = i * sectionHeight / scale;
//         const height = sectionHeight / scale;

//         // Draw section background
//         ctx.fillStyle = colors[i];
//         ctx.fillRect(0, y, canvas.width / scale, height);

//         // Draw section label
//         ctx.fillStyle = 'rgba(0,0,0,0.7)';
//         ctx.font = `${16 / scale}px Arial`;
//         ctx.fillText(sections[i], 10, y + 30 / scale);
//     }
// }

// // Draw nodes
// function drawNodes() {
//     pipelineData.nodes.forEach(node => {
//         // Skip query nodes if toggle is off
//         if (node.type === 'query' && !showQueries) return;

//         ctx.strokeStyle = 'black';
//         ctx.lineWidth = 1 / scale;
        
//         // Use squares for data sources, circles for process nodes
//         if (node.type === 'source') {
//             // Draw as square for data sources
//             ctx.fillStyle = getNodeColor(node.type);
//             ctx.beginPath();
//             ctx.rect(node.x, node.y, node.width, node.height);
//             ctx.fill();
//             ctx.stroke();
            
//             // Show name for data sources
//             ctx.fillStyle = 'black';
//             ctx.font = `${12 / scale}px Arial`;
//             ctx.textAlign = 'center';
//             ctx.fillText(node.name, node.x + node.width / 2, node.y + node.height / 2);
//         } else {
//             // Draw as circle for process nodes
//             const radius = node.width / 2;
//             ctx.fillStyle = getNodeColor(node.type);
//             ctx.beginPath();
//             ctx.arc(node.x + radius, node.y + radius, radius, 0, Math.PI * 2);
//             ctx.fill();
//             ctx.stroke();
            
//             // Show abbreviated name for process nodes
//             ctx.fillStyle = 'black';
//             ctx.font = `${10 / scale}px Arial`;
//             ctx.textAlign = 'center';
//             ctx.fillText(node.id, node.x + radius, node.y + radius);
//         }
        
//         // Highlight selected node
//         if (selectedNode && node.id === selectedNode.id) {
//             ctx.strokeStyle = 'yellow';
//             ctx.lineWidth = 3 / scale;
//             if (node.type === 'source') {
//                 ctx.strokeRect(node.x, node.y, node.width, node.height);
//             } else {
//                 const radius = node.width / 2;
//                 ctx.beginPath();
//                 ctx.arc(node.x + radius, node.y + radius, radius, 0, Math.PI * 2);
//                 ctx.stroke();
//             }
//         }
//     });
// }

// // Draw edges between nodes
// function drawEdges() {
//     pipelineData.edges.forEach(edge => {
//         const fromNode = pipelineData.nodes.find(n => n.id === edge.from);
//         const toNode = pipelineData.nodes.find(n => n.id === edge.to);

//         // Skip edges to/from query nodes if toggle is off
//         if ((fromNode.type === 'query' || toNode.type === 'query') && !showQueries) return;

//         ctx.strokeStyle = 'rgba(100,100,100,0.7)';
//         ctx.lineWidth = 2 / scale;

//         // Calculate start and end points
//         let startX, startY, endX, endY;
        
//         if (fromNode.type === 'source') {
//             startX = fromNode.x + fromNode.width / 2;
//             startY = fromNode.y + fromNode.height;
//         } else {
//             const radius = fromNode.width / 2;
//             startX = fromNode.x + radius;
//             startY = fromNode.y + radius;
            
//             // Adjust based on direction to next node
//             const angle = Math.atan2(toNode.y - fromNode.y, toNode.x - fromNode.x);
//             startX += Math.cos(angle) * radius;
//             startY += Math.sin(angle) * radius;
//         }
        
//         if (toNode.type === 'source') {
//             endX = toNode.x + toNode.width / 2;
//             endY = toNode.y;
//         } else {
//             const radius = toNode.width / 2;
//             endX = toNode.x + radius;
//             endY = toNode.y + radius;
            
//             // Adjust based on direction from previous node
//             const angle = Math.atan2(toNode.y - fromNode.y, toNode.x - fromNode.x);
//             endX -= Math.cos(angle) * radius;
//             endY -= Math.sin(angle) * radius;
//         }

//         // Draw the line
//         ctx.beginPath();
//         ctx.moveTo(startX, startY);
//         ctx.lineTo(endX, endY);
//         ctx.stroke();

//         // Draw arrow
//         const arrowSize = 5 / scale;
//         const angle = Math.atan2(endY - startY, endX - startX);

//         ctx.beginPath();
//         ctx.moveTo(endX, endY);
//         ctx.lineTo(
//             endX - arrowSize * Math.cos(angle - Math.PI / 6),
//             endY - arrowSize * Math.sin(angle - Math.PI / 6)
//         );
//         ctx.lineTo(
//             endX - arrowSize * Math.cos(angle + Math.PI / 6),
//             endY - arrowSize * Math.sin(angle + Math.PI / 6)
//         );
//         ctx.closePath();
//         ctx.fill();
//     });
// }

// // Get node color based on type
// function getNodeColor(type) {
//     switch (type) {
//         case 'source': return 'rgba(80,180,80,0.8)';
//         case 'extraction': return 'rgba(120,200,120,0.8)';
//         case 'transformation': return 'rgba(100,100,200,0.8)';
//         case 'load': return 'rgba(200,100,100,0.8)';
//         case 'query': return 'rgba(200,200,100,0.8)';
//         default: return 'rgba(150,150,150,0.8)';
//     }
// }

// // Check if mouse is over a node
// function getNodeAtPosition(x, y) {
//     // Adjust coordinates for pan and zoom
//     x = (x - panOffset.x) / scale;
//     y = (y - panOffset.y) / scale;

//     for (let i = pipelineData.nodes.length - 1; i >= 0; i--) {
//         const node = pipelineData.nodes[i];

//         // Skip query nodes if toggle is off
//         if (node.type === 'query' && !showQueries) continue;

//         if (node.type === 'source') {
//             // Square hit detection for data sources
//             if (x >= node.x && x <= node.x + node.width &&
//                 y >= node.y && y <= node.y + node.height) {
//                 return node;
//             }
//         } else {
//             // Circle hit detection for process nodes
//             const radius = node.width / 2;
//             const centerX = node.x + radius;
//             const centerY = node.y + radius;
//             const distance = Math.sqrt(Math.pow(x - centerX, 2) + Math.pow(y - centerY, 2));
            
//             if (distance <= radius) {
//                 return node;
//             }
//         }
//     }
//     return null;
// }

// // Render node details when a node is clicked
// function renderNodeDetails(node) {
//     // Show the node details container
//     const mainTimeline = document.querySelector('.main-timeline');
//     const mainEditor = document.querySelector('.main-editor');
    
//     mainTimeline.style.display = 'none';
//     mainEditor.style.display = 'block';
    
//     // Clear previous content
//     mainEditor.innerHTML = '';
    
//     // Create header with back button
//     const header = document.createElement('div');
//     header.className = 'editor-header';
    
//     const backButton = document.createElement('button');
//     backButton.innerHTML = '&larr; Back to Timeline';
//     backButton.className = 'btn btn-secondary';
//     backButton.onclick = () => {
//         mainEditor.style.display = 'none';
//         mainTimeline.style.display = 'block';
//     };
    
//     const title = document.createElement('h2');
//     title.textContent = `${node.name} (${node.id})`;
    
//     header.appendChild(backButton);
//     header.appendChild(title);
//     mainEditor.appendChild(header);
    
//     // If it's a node with SQL, show the code section
//     if (node.code) {
//         const codeSection = document.createElement('div');
//         codeSection.className = 'code-section';
        
//         const codeTitle = document.createElement('h3');
//         codeTitle.textContent = 'SQL Code';
        
//         const codeBlock = document.createElement('pre');
//         codeBlock.className = 'sql-code';
//         codeBlock.textContent = node.code;
        
//         codeSection.appendChild(codeTitle);
//         codeSection.appendChild(codeBlock);
//         mainEditor.appendChild(codeSection);
//     }
    
//     // If node has schema, show data table
//     if (node.schema) {
//         const dataSection = document.createElement('div');
//         dataSection.className = 'data-section';
        
//         const dataTitle = document.createElement('h3');
//         dataTitle.textContent = 'Data Preview';
        
//         const table = document.createElement('table');
//         table.className = 'data-table';
        
//         // Create header row
//         const thead = document.createElement('thead');
//         const headerRow = document.createElement('tr');
        
//         node.schema.forEach(column => {
//             const th = document.createElement('th');
//             th.innerHTML = `${column.name}<br><span class="column-type">${column.type}</span>`;
//             headerRow.appendChild(th);
//         });
        
//         thead.appendChild(headerRow);
//         table.appendChild(thead);
        
//         // Create data rows (generate fake data based on schema)
//         const tbody = document.createElement('tbody');
//         const rowCount = 10;
        
//         // Generate random data based on column types
//         const generateValue = (type) => {
//             switch(type) {
//                 case 'string':
//                     return `sample-${Math.random().toString(36).substring(2, 8)}`;
//                 case 'integer':
//                     return Math.floor(Math.random() * 1000);
//                 case 'float':
//                     return +(Math.random() * 100).toFixed(2);
//                 case 'boolean':
//                     return Math.random() > 0.5 ? 'true' : 'false';
//                 case 'date':
//                     const date = new Date();
//                     date.setDate(date.getDate() - Math.floor(Math.random() * 365));
//                     return date.toISOString().split('T')[0];
//                 default:
//                     return 'N/A';
//             }
//         };
        
//         for (let i = 0; i < rowCount; i++) {
//             const dataRow = document.createElement('tr');
            
//             node.schema.forEach(column => {
//                 const td = document.createElement('td');
//                 td.textContent = generateValue(column.type);
//                 dataRow.appendChild(td);
//             });
            
//             tbody.appendChild(dataRow);
//         }
        
//         table.appendChild(tbody);
//         dataSection.appendChild(dataTitle);
//         dataSection.appendChild(table);
//         mainEditor.appendChild(dataSection);
//     }
// }

// // Update sidebar with node configuration
// function updateSidebar(node) {
//     const sidebar = document.querySelector('.sidebar');
//     const sideHistory = document.getElementById('history');
    
//     sidebar.style.display = 'block';
//     document.querySelector('main').style.margin = '0 40px';
    
//     // Generate configuration info for sidebar
//     let configInfo = `# ${node.name} (${node.id})\n\n`;
//     configInfo += `**Type:** ${node.type}\n\n`;
//     configInfo += `**Details:** ${node.details}\n\n`;
    
//     if (node.type === 'load' && node.stats) {
//         configInfo += `## Statistics\n\n`;
//         configInfo += `- Data Stored: ${node.stats.dataStored}\n`;
//         configInfo += `- Last Refresh: ${new Date(node.stats.lastRefresh).toLocaleString()}\n`;
//         configInfo += `- Avg Refresh Time: ${node.stats.avgRefreshTime}\n`;
//         configInfo += `- Query Count: ${node.stats.queryCount.toLocaleString()}\n`;
//         configInfo += `- Refresh Frequency: ${node.stats.refreshFrequency}\n\n`;
//     }
    
//     if (node.schema) {
//         configInfo += `## Schema\n\n`;
//         node.schema.forEach(col => {
//             configInfo += `- ${col.name} (${col.type})\n`;
//         });
//     }
    
//     sideHistory.value = configInfo;
// }

// // Populate the Refresh History section
// function populateRefreshes() {
//     const refreshesTab = document.querySelector('[data-tab="1"]');
    
//     // Clear previous content
//     refreshesTab.innerHTML = '';
    
//     // Create all refreshes table
//     const allRefreshesSection = document.createElement('div');
//     allRefreshesSection.innerHTML = '<h3>All ETL Refreshes</h3>';
    
//     const allRefreshesTable = document.createElement('table');
//     allRefreshesTable.className = 'data-table';
//     allRefreshesTable.innerHTML = `
//         <thead>
//             <tr>
//                 <th>Data Source</th>
//                 <th>Refresh Time</th>
//                 <th>Type</th>
//                 <th>Duration</th>
//                 <th>Status</th>
//             </tr>
//         </thead>
//         <tbody>
//             ${fakeData.refreshHistory.map(refresh => `
//                 <tr class="${refresh.status === 'Failed' ? 'failed-row' : ''}">
//                     <td>${refresh.datasource}</td>
//                     <td>${new Date(refresh.refreshTime).toLocaleString()}</td>
//                     <td>${refresh.refreshType}</td>
//                     <td>${refresh.duration}</td>
//                     <td class="status-${refresh.status.toLowerCase()}">${refresh.status}</td>
//                 </tr>
//             `).join('')}
//         </tbody>
//     `;
    
//     allRefreshesSection.appendChild(allRefreshesTable);
//     refreshesTab.appendChild(allRefreshesSection);
    
//     // Create failed processes table
//     const failedSection = document.createElement('div');
//     failedSection.innerHTML = '<h3>Failed ETL Processes</h3>';
    
//     const failedTable = document.createElement('table');
//     failedTable.className = 'data-table';
//     failedTable.innerHTML = `
//         <thead>
//             <tr>
//                 <th>Data Source</th>
//                 <th>Failed Step</th>
//                 <th>Time</th>
//                 <th>Error Message</th>
//             </tr>
//         </thead>
//         <tbody>
//             ${fakeData.failedProcesses.map(failure => `
//                 <tr>
//                     <td>${failure.datasource}</td>
//                     <td>${failure.step}</td>
//                     <td>${new Date(failure.crashTime).toLocaleString()}</td>
//                     <td class="error-message">${failure.errorMessage}</td>
//                 </tr>
//             `).join('')}
//         </tbody>
//     `;
    
//     failedSection.appendChild(failedTable);
//     refreshesTab.appendChild(failedSection);
// }

// // Populate the Performance section
// function populatePerformance() {
//     const performanceTab = document.querySelector('[data-tab="2"]');
    
//     // Clear previous content
//     performanceTab.innerHTML = '';
    
//     // Create metric cards
//     const metrics = fakeData.performanceMetrics;
//     const metricsSection = document.createElement('div');
//     metricsSection.className = 'metric-cards';
    
//     metricsSection.innerHTML = `
//         <div class="metric-card">
//             <h3>Avg Throughput</h3>
//             <div class="metric-value">${metrics.avgThroughput}</div>
//         </div>
//         <div class="metric-card">
//             <h3>Avg Latency</h3>
//             <div class="metric-value">${metrics.avgLatency}</div>
//         </div>
//         <div class="metric-card">
//             <h3>Error Rate</h3>
//             <div class="metric-value">${metrics.errorRate}</div>
//         </div>
//         <div class="metric-card">
//             <h3>Total Data Stored</h3>
//             <div class="metric-value">${metrics.totalDataStored}</div>
//         </div>
//         <div class="metric-card">
//             <h3>Avg Pipeline Time</h3>
//             <div class="metric-value">${metrics.avgPipelineTime}</div>
//         </div>
//     `;
    
//     performanceTab.appendChild(metricsSection);
    
//     // Create serving layer table
//     const servingSection = document.createElement('div');
//     servingSection.innerHTML = '<h3>Serving Layer Statistics</h3>';
    
//     const servingTable = document.createElement('table');
//     servingTable.className = 'data-table';
//     servingTable.innerHTML = `
//         <thead>
//             <tr>
//                 <th>View Name</th>
//                 <th>Data Stored</th>
//                 <th>Last Refresh</th>
//                 <th>Avg Refresh Time</th>
//                 <th>Query Count</th>
//                 <th>Refresh Frequency</th>
//             </tr>
//         </thead>
//         <tbody>
//             ${fakeData.servingLayers.map(layer => `
//                 <tr>
//                     <td>${layer.name}</td>
//                     <td>${layer.dataStored}</td>
//                     <td>${new Date(layer.lastRefresh).toLocaleString()}</td>
//                     <td>${layer.avgRefreshTime}</td>
//                     <td>${layer.queryCount.toLocaleString()}</td>
//                     <td>${layer.refreshFrequency}</td>
//                 </tr>
//             `).join('')}
//         </tbody>
//     `;
    
//     servingSection.appendChild(servingTable);
//     performanceTab.appendChild(servingSection);
// }

// // Mouse move handler
// canvas.addEventListener('mousemove', (e) => {
//     const rect = canvas.getBoundingClientRect();
//     const mouseX = e.clientX - rect.left;
//     const mouseY = e.clientY - rect.top;

//     // Handle dragging
//     if (isDragging) {
//         const dx = mouseX - lastMousePos.x;
//         const dy = mouseY - lastMousePos.y;

//         panOffset.x += dx;
//         panOffset.y += dy;

//         // Limit panning
//         const maxPan = canvas.width * 0.5;
//         panOffset.x = Math.max(Math.min(panOffset.x, maxPan), -maxPan);
//         panOffset.y = Math.max(Math.min(panOffset.y, maxPan), -maxPan);

//         lastMousePos = { x: mouseX, y: mouseY };
//         drawPipeline();
//     } else {
//         // Check for hover
//         const node = getNodeAtPosition(mouseX, mouseY);

//         if (node) {
//             canvas.style.cursor = 'pointer';
//             // Show tooltip
//             tooltip.style.display = 'block';
//             tooltip.style.left = (e.clientX + 10) + 'px';
//             tooltip.style.top = (e.clientY + 10) + 'px';
//             tooltip.textContent = `${node.name}: ${node.details}`;
//         } else {
//             canvas.style.cursor = 'default';
//             tooltip.style.display = 'none';
//         }
//     }
// });

// // Mouse down handler
// canvas.addEventListener('mousedown', (e) => {
//     const rect = canvas.getBoundingClientRect();
//     const mouseX = e.clientX - rect.left;
//     const mouseY = e.clientY - rect.top;

//     const node = getNodeAtPosition(mouseX, mouseY);

//     if (node) {
//         selectedNode = node;
//         // Show node details in the main view
//         renderNodeDetails(node);
//         // Update sidebar with node configuration
//         updateSidebar(node);
//     } else {
//         // Start panning
//         isDragging = true;
//         lastMousePos = { x: mouseX, y: mouseY };
//     }
// });

// // Mouse up handler
// canvas.addEventListener('mouseup', () => {
//     isDragging = false;
// });

// // Mouse out handler
// canvas.addEventListener('mouseout', () => {
//     isDragging = false;
//     tooltip.style.display = 'none';
// });

// // Mouse wheel handler for zoom
// canvas.addEventListener('wheel', (e) => {
//     e.preventDefault();

//     const rect = canvas.getBoundingClientRect();
//     const mouseX = e.clientX - rect.left;
//     const mouseY = e.clientY - rect.top;

//     // Calculate zoom factor
//     const zoom = e.deltaY < 0 ? 1.1 : 0.9;

//     // Apply zoom limits
//     const newScale = Math.max(0.5, Math.min(2.0, scale * zoom));

//     if (newScale !== scale) {
//         // Zoom centered on mouse position
//         const scaleRatio = newScale / scale;
//         panOffset.x = mouseX - scaleRatio * (mouseX - panOffset.x);
//         panOffset.y = mouseY - scaleRatio * (mouseY - panOffset.y);

//         scale = newScale;
//         drawPipeline();
//     }
// });

// // Query toggle handler
// queryToggle.addEventListener('change', (e) => {
//     showQueries = e.target.checked;
//     drawPipeline();
// });

// // Add CSS for the nodes detail view and performance metrics
// const style = document.createElement('style');
// style.textContent = `
//     .editor-header {
//         display: flex;
//         align-items: center;
//         margin-bottom: 20px;
//     }
    
//     .editor-header h2 {
//         margin-left: 20px;
//     }
    
//     .code-section {
//         margin-bottom: 20px;
//     }
    
//     .sql-code {
//         background-color: #f5f5f5;
//         padding: 15px;
//         border-radius: 5px;
//         overflow-x: auto;
//         font-family: monospace;
//     }
    
//     .data-table {
//         width: 100%;
//         border-collapse: collapse;
//         margin-top: 10px;
//     }
    
//     .data-table th, .data-table td {
//         border: 1px solid #ddd;
//         padding: 8px;
//         text-align: left;
//     }
    
//     .data-table th {
//         background-color: #f2f2f2;
//     }
    
//     .column-type {
//         font-size: 0.8em;
//         color: #666;
//     }
    
//     .metric-cards {
//         display: flex;
//         flex-wrap: wrap;
//         gap: 20px;
//         margin-bottom: 30px;
//     }
    
//     .metric-card {
//         background-color: #f5f5f5;
//         border-radius: 8px;
//         padding: 15px;
//         min-width: 180px;
//         flex: 1;
//         box-shadow: 0 2px 4px rgba(0,0,0,0.1);
//     }
    
//     .metric-card h3 {
//         margin-top: 0;
//         color: #555;
//         font-size: 0.9em;
//     }
    
//     .metric-value {
//         font-size: 1.8em;
//         font-weight: bold;
//         color: #333;
//     }
    
//     .failed-row {
//         background-color: rgba(255, 200, 200, 0.3);
//     }
    
//     .status-success {
//         color: green;
//         font-weight: bold;
//     }
    
//     .status-failed {
//         color: red;
//         font-weight: bold;
//     }
    
//     .error-message {
//         color: #d32f2f;
//         max-width: 400px;
//     }
// `;
// document.head.appendChild(style);

// // Setup tab navigation
// document.querySelectorAll('.tab-navigation button').forEach(button => {
//     button.addEventListener('click', () => {
//         const tabId = button.getAttribute('data-tab');
        
//         // Remove active class from all buttons and tabs
//         document.querySelectorAll('.tab-navigation button').forEach(btn => {
//             btn.classList.remove('active');
//         });
//         document.querySelectorAll('.tab-content').forEach(tab => {
//             tab.classList.remove('active');
//         });
        
//         // Add active class to current button and tab
//         button.classList.add('active');
//         document.querySelector(`.tab-content[data-tab="${tabId}"]`).classList.add('active');
//     });
// });

// // Initialize tab content
// populateRefreshes();
// populatePerformance();

// // Initial setup
// document.querySelector('.tab-navigation button[data-tab="1"]').click();

// // Initial draw
// drawPipeline();

// // Handle window resize
// window.addEventListener('resize', () => {
//     canvas.width = canvas.offsetWidth;
//     canvas.height = canvas.offsetHeight;
//     drawPipeline();
// });
