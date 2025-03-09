import Chart from 'chart.js/auto';
import { Grid } from 'gridjs';

// Chart.js initialization
const ctx = (document.getElementById('myChart') as HTMLCanvasElement).getContext('2d');
if (ctx) {
  new Chart(ctx, {
    type: 'bar',
    data: {
      labels: ['January', 'February', 'March', 'April'],
      datasets: [{
        label: 'Production',
        data: [10, 20, 30, 40],
        backgroundColor: 'rgba(0, 123, 255, 0.5)',
        borderColor: 'rgba(0, 123, 255, 1)',
        borderWidth: 1
      }]
    },
    options: {
      responsive: true,
      plugins: {
        legend: {
          display: true
        }
      },
      scales: {
        y: {
          beginAtZero: true
        }
      }
    }
  });
}

// Grid.js table initialization
new Grid({
  columns: ["Machine", "Shift", "Amount Produced"],
  data: [
    ["Machine 1", "Shift 1", 100],
    ["Machine 1", "Shift 2", 150],
    ["Machine 2", "Shift 1", 120],
    ["Machine 2", "Shift 2", 170]
  ],
  pagination: true,
  search: true,
  sort: true
}).render(document.getElementById("table-wrapper")as HTMLElement);