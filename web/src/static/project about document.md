
Main Points
--
- What the project is about and it's aim?
- How the theory works?
- Technology used?
- Optimisations Used
- Things left to do

### Introduction


Purpose of this project is mainly for my own learning of handling big data and data analysis in order to make a scalable, reliable and maintainable data driven application.

#### Project Aims

I want to build an application that can data from wide range sources, and easily be able to query it efficiently to get both live and and historical data with minimal latency.

Businesses may have a lot of data from multiple different types of vendors, software’s, and systems. They are interacted with in different, export and import in different ways and might not be able entirely useful or used to the best of their ability until they are combined with another system that again, might have an entirely different eco system.

Aimed towards industries like logistics, transportation, manufacturing, finances, where tracking live data and analytics across large complex systems are necessary.

### How it works

Main concepts come from the **ETL** Process (Extract Transform and Load) **Lambda Architecture** by Nathan Marz as described in his book "Principles and best practices of scalable real-time data systems" in 2015.

![[Pasted image 20250204145944.webp]]

The main goal is for data analysts to query **any** data without needing to worry about timeliness, latency, accuracy, performance etc. and have a high quality reliable overview of data

Has 2 Main layers, **Speed Layer** and **Batch Layer**:

**Batch Layer**:  is responsible for handlining querying  historical data by storing all data sources into one data ware house, and then analysts build queries against that data that refresh and selected intervals. (This is querying section is what the **serving layer** is used for). **Views stored in a Postgres Database**
**Speed Layer**: New data past the batch layer refresh interval in order to get live data. Views stored in a **Redis cache**

Cons:
- **2 Separate processing layer** engines which are hard to maintain. *Almost similar* code but enough differences that it becomes difficult to abstract and errors typical duplicate in both, making it very hard to debug.
- **Needs Incremental refreshes**, due to full refreshes being too intensive repeatedly. Makes the batch layer much more complicated. 
- **Speed Layer is incredibly unreliable** since it doesn't last as long as the batch layer, meaning if a duplicate message is received hours later, it won't be able to tell it's already received it since having forgotten. This damages the accuracy of the speed layer.
- **Combining results from both layers** is incredibly difficult and requires some difficult computation.
- **Speed layer is not as intensive**, since compared to the batch layer, it can only take **live** data sources so streams from sensors. Getting live data from *most* sources is very difficult and requires a lot of work

### Test data

Just for the testing of this application, I need example scenario to create test data for, so having example reports, queries, test data,
data sources etc.

Hence **MyCPU**, a fictional single floored factory that go from raw components to assembling and shipping out complete CPU's
This is how the factory floor looks like:
![[IMG_20250211_185705-1.webp|409]]
Reports for this application:

- Daily floor production summary (Map contains live data travelling, bottle neck showing heat map concentration of parts at different locations over different spans of times, tool tips parts produced, different rates, other KPI metrics)
- workforce utilisation report (attendance for each department, shift efficient, tasks per department, different reasons for not attending.)
- Energy consumption report (power usage across the different machines, Again showing a concertation against all of them via a heat map)
- Quality control for parts (Amount of parts that have failed vs haven't. Amount reworked, amount pending rework, MTTR, Average Lifespan of tools, Failure of parts against different departments and different stations)




| **Report Name**               | **Purpose of Report**                                                                                                                                             | **Queries for Report**                                                                                | **Sources for this Data**                                  |
| ----------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------- | ---------------------------------------------------------- |
| Daily floor productionsummary | Main page to showcase all of the mainfeatures of the application, including showing live updates of data as well as total summary across different ranges of time | - Show daily production totals<br>- Display shift<br>- Wise performance<br>- Show time-based analysis | - Worker Clock in Sheet<br>- Tasks<br>- Oracle Database    |
| Workforce utilisation report  | Page dedicated to showing theinformation of the workers of thiscompany.                                                                                           | - Employee attendance<br>- Utilisation rate per worker<br>- Work hours per task                       | - Worker Time Logs<br>- HR Database                        |
| Energy consumption report     | Another page to demonstrate live data                                                                                                                             | - Total energy consumption<br>- Energy usage by department<br>- Efficiency of energy use              | - Energy Meter Data<br>- Building Energy Management System |
| Quality control for parts     | Tool for the quality department inorder to view the maintenance of eachof the parts                                                                               | - Quality status of parts<br>- Maintenance records<br>- Inspection results                            | - Quality Control Database<br>- Maintenance Logs           |

Data sources needed.


| **Data Source Name**       | **What it Contains**                                                                                           | **Type of Data** | **Refresh**                 |
| -------------------------- | -------------------------------------------------------------------------------------------------------------- | ---------------- | --------------------------- |
| Worker Clock In Sheet      | Attendance sheet by department, statingwho they are, what shift they are on andreason for absence if not here  | Excel            | 3 times a day               |
| Energy ConsumptionReport   | Energy Consumption Report against each Machine                                                                 | Excel/CSV        | Daily                       |
| Heat Sensor                | Sensor data from measuring Faulty Parts differentsections of the factory floor                                 | Kafka            | Live                        |
| Live Machining Data        | Data travelling at each location, stating where it's going,any error codes from station it's leaving           | Kafka + MSSQL    | Live + archive storeof data |
| Repair and Purchase Orders | Orders for incoming and outgoing parts, Severity of faults, order confirmation, request and received           | Postgres         | Up to 50 times a day        |
| Tool Storage inventory     | Tools storage so inventory of tools being taken in and outof storage                                           | Rest API         | 30 mins                     |
| Tool Master List           | Master list stating all of the information for tools, needed,like lifespan, IDs, usage cases, from vendor etc. | Excel/CSV        | Never                       |
| Factory Floor Mappings     | Maps of factory layout, workstations, and machine locations                                                    | JSON             | Never                       |


There is a script constantly running to create, add delete and cause errors in order to simulate this data across different data sources 
### System Architecture

Image for the layout of my architecture,

- Explanation of how data is combined

![[IMG_20250211_184544.webp|335]]

### Optimisations

##### Reading from sources live
- Reading from databases and static files in a number of different ways depending on the source:

| **Data Source**                              | **Native Change Tracking?**               | **Can Track Appended / Updated Data?** | **Common Techniques**                   |
| -------------------------------------------- | ----------------------------------------- | -------------------------------------- | --------------------------------------- |
| **Databases (SQL, NoSQL)**                   | Replication Logs (CDC, WAL, Binlog, etc.) | Yes                                    | CDC, Binlogs, WAL, Triggers             |
| **CSV / Text Files**                         | No built-in tracking                      | Only if appending                      | File size checks, Line offsets, Hashing |
| **Excel Files (XLSX, CSV)**                  | No built-in tracking                      | Appending possible                     | Versioning, Metadata timestamps         |
| **XML / JSON Files**                         | No built-in tracking                      | Appending possible                     | Diffing, Last Modified Time             |
| **Parquet / Avro** (Big Data Formats)        | Some support for metadata                 | Yes                                    | Partitioning, Append-Only Logs          |
| **Message Queues (Kafka, Pulsar, RabbitMQ)** | Built-in offset tracking                  | Yes                                    | Kafka Offsets, Consumer State           |
So mainly using replication logs for reading live data using **Debezium**

For static files, for both incremental batches and reading from live data sources, it can depend:

- You can check if it has even refreshed, via the **modified time date** and version of the actual file, and if it hasn't it can been used regularly.
- For logs, or files expected to only append, can keep a **line offset** and hash the data up to that point, so if the file updates and you hash that same amount of data, that data has changed and therefore anything appended is new that can be added (of course if the size of the file is the same, this is unnecessary)


#### **Redis Cache** vs **Postgres**

Each real time view will have 


##### Combining query data from 




| **Query Type**                                                            | **Complexity** | **Easy to Update?** | **How It's Handled**                                                         |
| ------------------------------------------------------------------------- | -------------- | ------------------- | ---------------------------------------------------------------------------- |
| **Simple Aggregations** (SUM, COUNT, AVG, MIN, MAX)                       | Low            | Easy                | Uses **incremental updates** (just update a counter or sum)                  |
| **Sliding & Tumbling Windows** (e.g., "Count last 5 minutes of data")     | Medium         | Kinda               | Uses **windowed state storage**, discarding old data when the window expires |
| **Joins Between Streams** (e.g., "Match transactions with user profiles") | High           | Harder              | Stores state and waits for matching data (can require a lot of memory)       |




- 