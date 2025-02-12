import React, { useEffect, useRef, useState } from 'react';
import * as d3 from 'd3';
import Papa from 'papaparse';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';

const RegulatoryDashboard = () => {
  const [agencyData, setAgencyData] = useState([]);
  const [totalData, setTotalData] = useState([]);
  const [growthData, setGrowthData] = useState([]);
  const [selectedYear, setSelectedYear] = useState(null);
  const [yearRange, setYearRange] = useState({ min: 2017, max: 2025 });
  const [dimensions, setDimensions] = useState({ width: 1800, height: 1200 }); // Increased dimensions
  const svgRef = useRef();
const simulation = useRef(null);
const nodePositions = useRef(new Map()); // Store positions for each agency

  useEffect(() => {
    const loadData = async () => {
      try {
        const [agencyResponse, totalResponse] = await Promise.all([
          fetch('/titles_all.csv'),
          fetch('/titles_totals.csv')
        ]);
        
        const [agencyText, totalText] = await Promise.all([
          agencyResponse.text(),
          totalResponse.text()
        ]);
        
        Papa.parse(agencyText, {
          header: false,
          skipEmptyLines: true,
          complete: (results) => {
            const processedData = results.data
              .map(row => ({
                year: parseInt(row[3]),
                agency: row[2],
                fullName: row[1],
                displayName: row[2] || row[1], // Fallback to full name if no short name
                wordCount: parseInt(row[4])
              }))
              .filter(item => !isNaN(item.year) && !isNaN(item.wordCount));

            const agencyYearlyGrowth = d3.group(processedData, d => d.agency);
            const growthData = Array.from(agencyYearlyGrowth).map(([agency, data]) => {
              const sortedData = data.sort((a, b) => a.year - b.year);
              const growth = sortedData.map((d, i) => {
                if (i === 0) return 0;
                const prevCount = sortedData[i - 1].wordCount;
                return ((d.wordCount - prevCount) / prevCount) * 100;
              });
              
              return {
                agency,
                averageGrowth: d3.mean(growth),
                totalGrowth: ((sortedData[sortedData.length - 1].wordCount - sortedData[0].wordCount) / sortedData[0].wordCount) * 100
              };
            });

            setGrowthData(growthData);
            
            const years = processedData.map(d => d.year);
            setYearRange({
              min: Math.min(...years),
              max: Math.max(...years)
            });
            setSelectedYear(Math.min(...years));
            setAgencyData(processedData);
          }
        });

        Papa.parse(totalText, {
          header: true,
          skipEmptyLines: true,
          complete: (results) => {
            const processedTotals = results.data
              .map(row => ({
                year: parseInt(row.Year),
                totalWords: parseInt(row.Amount)
              }))
              .filter(item => !isNaN(item.year) && !isNaN(item.totalWords))
              .sort((a, b) => a.year - b.year);

            processedTotals.forEach((d, i) => {
              if (i > 0) {
                const prevYear = processedTotals[i - 1];
                d.yearOverYearChange = ((d.totalWords - prevYear.totalWords) / prevYear.totalWords) * 100;
              } else {
                d.yearOverYearChange = 0;
              }
            });

            setTotalData(processedTotals);
          }
        });

      } catch (error) {
        console.error('Error reading files:', error);
      }
    };
    loadData();
  }, []);

  useEffect(() => {
    if (!agencyData.length || !selectedYear) return;

    const svg = d3.select(svgRef.current);
    svg.selectAll("*").remove();

    const yearData = agencyData.filter(d => d.year === selectedYear);
    
    const sizeScale = d3.scaleSqrt()
      .domain([0, d3.max(agencyData, d => d.wordCount)])
      .range([30, 120]); // Increased bubble size range

    const colorScale = d3.scaleOrdinal(d3.schemeSet2);

    const container = svg.append("g")
      .attr("transform", `translate(${dimensions.width / 2}, ${dimensions.height / 2})`);

    simulation.current = d3.forceSimulation(yearData)
      .force("center", d3.forceCenter(0, 0).strength(1)) // Increased center force
      .force("charge", d3.forceManyBody().strength(10)) // Added negative charge for repulsion
      .force("collide", d3.forceCollide().radius(d => sizeScale(d.wordCount) + 2).strength(0.8))
      .velocityDecay(0.3) // Reduced decay to allow more movement
      .on("tick", () => {
        bubbleGroups.attr("transform", d => {
          const radius = sizeScale(d.wordCount);
          d.x = Math.max(-dimensions.width + radius, Math.min(dimensions.width - radius, d.x));
          d.y = Math.max(-dimensions.height + radius, Math.min(dimensions.height - radius, d.y));
          return `translate(${d.x},${d.y})`;
        });
      });

    const bubbleGroups = container.selectAll("g")
      .data(yearData)
      .join("g")
      .attr("class", "bubble-group");

    bubbleGroups.append("circle")
      .attr("r", d => sizeScale(d.wordCount))
      .style("fill", d => colorScale(d.agency))
      .style("fill-opacity", 0.8) // Added slight transparency
      .style("stroke", "white")
      .style("stroke-width", 2)
      .style("cursor", "pointer")
      .transition()
      .duration(400)
      .ease(d3.easeCubicInOut)
      .attrTween("r", function(d) {
        const i = d3.interpolate(0, sizeScale(d.wordCount));
        return t => i(t);
      });

    // Background for text
    bubbleGroups.append("text")
      .attr("class", "agency-label-bg")
      .attr("text-anchor", "middle")
      .attr("dy", "-0.5em")
      .style("fill", "white")
      .style("stroke", "white")
      .style("stroke-width", "6px")
      .style("stroke-opacity", "0.8")
      .style("paint-order", "stroke")
      .text(d => d.displayName)
      .style("font-size", d => `${Math.min(sizeScale(d.wordCount) / 4, 16)}px`);

    // Actual text
    bubbleGroups.append("text")
      .attr("class", "agency-label")
      .attr("text-anchor", "middle")
      .attr("dy", "-0.5em")
      .style("fill", "#333")
      .style("font-weight", "bold")
      .style("pointer-events", "none")
      .text(d => d.displayName)
      .style("font-size", d => `${Math.min(sizeScale(d.wordCount) / 4, 16)}px`);

    // Word count with background
    bubbleGroups.append("text")
      .attr("class", "count-label-bg")
      .attr("text-anchor", "middle")
      .attr("dy", "1em")
      .style("fill", "white")
      .style("stroke", "white")
      .style("stroke-width", "6px")
      .style("stroke-opacity", "0.8")
      .style("paint-order", "stroke")
      .text(d => d.wordCount.toLocaleString())
      .style("font-size", d => `${Math.min(sizeScale(d.wordCount) / 5, 14)}px`);

    bubbleGroups.append("text")
      .attr("class", "count-label")
      .attr("text-anchor", "middle")
      .attr("dy", "1em")
      .style("fill", "#333")
      .style("pointer-events", "none")
      .text(d => d.wordCount.toLocaleString())
      .style("font-size", d => `${Math.min(sizeScale(d.wordCount) / 5, 14)}px`);

    bubbleGroups.on("mouseover", function(event, d) {
      d3.select(this).select("circle")
        .transition()
        .duration(200)
        .style("stroke", "#333")
        .style("stroke-width", 3);
    })
    .on("mouseout", function(event, d) {
      d3.select(this).select("circle")
        .transition()
        .duration(200)
        .style("stroke", "white")
        .style("stroke-width", 2);
    });

    return () => {
      if (simulation.current) simulation.current.stop();
    };
  }, [agencyData, selectedYear, dimensions]);

  return (
    <div>
      <div className="w-full mx-auto p-8 bg-white rounded-lg shadow-lg space-y-8 text-gray-900">
        <h1 className="text-4xl font-bold text-center mb-8 text-gray-900">Federal Regulations Analysis Dashboard</h1>
        <div>
            <h2 className="text-2xl font-bold text-center mb-4 text-gray-900">Annual Change in Total Word Count</h2>
            <div className="h-80">
                <ResponsiveContainer width="80%" height="100%" className="mx-auto">
                    <LineChart data={totalData}>
                        <CartesianGrid strokeDasharray="3 3" />
                        <XAxis 
                            dataKey="year"
                            type="number"
                            domain={['auto', 'auto']}
                            tickCount={totalData.length}
                        />
                        <YAxis
                            tickFormatter={(value) => `${value.toFixed(1)}%`}
                            domain={[0, 'auto']}
                        />
                        <Tooltip 
                            formatter={(value) => `${value?.toFixed(2)}%`}
                            labelFormatter={(value) => `Year: ${value}`}
                        />
                        <Line
                            type="monotone"
                            dataKey="yearOverYearChange"
                            stroke="#4ade80"
                            strokeWidth={2}
                            name="Year-over-Year Change"
                            dot={{ r: 4 }}
                            activeDot={{ r: 6 }}
                        />
                    </LineChart>
                </ResponsiveContainer>
            </div>
        </div>
        
        <div className="mb-12">
            <h2 className="text-2xl font-bold text-center mb-4 text-gray-900">Agency Word Count Distribution ({selectedYear})</h2>
            {selectedYear && (
                <div className="flex items-center justify-center gap-4 mb-4">
                    <span className="text-sm font-medium text-gray-900">Year:</span>
                    <input
                        type="range"
                        min={yearRange.min}
                        max={yearRange.max}
                        value={selectedYear}
                        onChange={(e) => setSelectedYear(parseInt(e.target.value))}
                        className="w-64"
                    />
                    <span className="text-sm font-medium text-gray-900">{selectedYear}</span>
                </div>
            )}
            <div className="flex justify-center">
                <svg
                    ref={svgRef}
                    width={dimensions.width}
                    height={dimensions.height}
                    className="bg-slate-50 rounded-lg"
                />
            </div>
        </div>
        <div className="mt-8 overflow-x-auto">
          <h3 className="text-xl font-bold mb-4">Raw Agency Data</h3>
          <table className="min-w-full text-sm border-collapse border border-gray-300">
            <thead className="bg-gray-100">
              <tr>
                <th className="border border-gray-300 px-2 py-1">Year</th>
                <th className="border border-gray-300 px-2 py-1">Agency</th>
                <th className="border border-gray-300 px-2 py-1">Full Name</th>
                <th className="border border-gray-300 px-2 py-1">Display Name</th>
                <th className="border border-gray-300 px-2 py-1">Word Count</th>
              </tr>
            </thead>
            <tbody>
              {agencyData.map((row, idx) => (
                <tr key={idx}>
                  <td className="border border-gray-300 px-2 py-1">{row.year}</td>
                  <td className="border border-gray-300 px-2 py-1">{row.agency}</td>
                  <td className="border border-gray-300 px-2 py-1">{row.fullName}</td>
                  <td className="border border-gray-300 px-2 py-1">{row.displayName}</td>
                  <td className="border border-gray-300 px-2 py-1">{row.wordCount}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};

const formatYAxis = (value) => {
  if (value >= 1000000) {
    return `${(value / 1000000).toFixed(1)}M`;
  }
  if (value >= 1000) {
    return `${(value / 1000).toFixed(0)}K`;
  }
  return value;
};

export default RegulatoryDashboard;
const TableView = ({ agencyData, selectedYear }) => {
    const tableData = React.useMemo(() => {
        if (!agencyData || !selectedYear) return [];
        
        return agencyData
            .filter(d => d.year === selectedYear)
            .sort((a, b) => b.wordCount - a.wordCount)
            .map(d => ({
                agency: d.displayName,
                fullName: d.fullName,
                wordCount: d.wordCount,
            }));
    }, [agencyData, selectedYear]);

    return (
        <div className="mt-8">
            <h2 className="text-2xl font-bold text-center mb-4 text-gray-900">Agency Word Count Table ({selectedYear})</h2>
            <div className="overflow-x-auto">
                <table className="min-w-full table-auto">
                    <thead>
                        <tr className="bg-gray-100">
                            <th className="px-4 py-2 text-left">Agency</th>
                            <th className="px-4 py-2 text-left">Full Name</th>
                            <th className="px-4 py-2 text-right">Word Count</th>
                        </tr>
                    </thead>
                    <tbody>
                        {tableData.map((row, i) => (
                            <tr key={row.agency} className={i % 2 === 0 ? 'bg-white' : 'bg-gray-50'}>
                                <td className="px-4 py-2 font-medium">{row.agency}</td>
                                <td className="px-4 py-2">{row.fullName}</td>
                                <td className="px-4 py-2 text-right">{row.wordCount.toLocaleString()}</td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
        </div>
    );
};