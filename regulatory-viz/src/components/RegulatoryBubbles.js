import React, { useEffect, useRef, useState } from 'react';
import * as d3 from 'd3';
import Papa from 'papaparse';

const RegulatoryBubbles = () => {
  const [data, setData] = useState([]);
  const [selectedYear, setSelectedYear] = useState(null);
  const [yearRange, setYearRange] = useState({ min: 2017, max: 2025 });
  const [dimensions, setDimensions] = useState({ width: 800, height: 600 });
  const svgRef = useRef();
  const simulation = useRef(null);

  // Load and process CSV data
  useEffect(() => {
    const loadData = async () => {
      try {
        const response = await fetch('/titles_all_nc.csv');
        const csvText = await response.text();
        
        Papa.parse(csvText, {
          header: false, // Since we know the exact column structure
          skipEmptyLines: true,
          complete: (results) => {
            // Transform the data into the required format
            const processedData = results.data.map(row => ({
              year: parseInt(row[3]), // Year is the 4th column
              agency: row[2], // Using ShortName for display
              fullName: row[1], // Keep DisplayName for tooltips
              wordCount: parseInt(row[4]) // TotalWords is the 5th column
            })).filter(item => !isNaN(item.year) && !isNaN(item.wordCount));

            // Calculate year range from data
            const years = processedData.map(d => d.year);
            setYearRange({
              min: Math.min(...years),
              max: Math.max(...years)
            });
            setSelectedYear(Math.min(...years)); // Set initial year
            setData(processedData);
          },
          error: (error) => {
            console.error('Error parsing CSV:', error);
          }
        });
      } catch (error) {
        console.error('Error reading file:', error);
      }
    };
    loadData();
  }, []);

  useEffect(() => {
    if (!data.length || !selectedYear) return;

    const svg = d3.select(svgRef.current);
    svg.selectAll("*").remove();

    const yearData = data.filter(d => d.year === selectedYear);
    
    const sizeScale = d3.scaleSqrt()
      .domain([0, d3.max(data, d => d.wordCount)])
      .range([20, 100]);

    const colorScale = d3.scaleOrdinal(d3.schemeSet2);

    const container = svg.append("g")
      .attr("transform", `translate(${dimensions.width / 2}, ${dimensions.height / 2})`);

    simulation.current = d3.forceSimulation(yearData)
      .force("center", d3.forceCenter(0, 0).strength(0.05))
      .force("x", d3.forceX().strength(0.08))
      .force("y", d3.forceY().strength(0.08))
      .force("charge", d3.forceManyBody().strength(5))
      .force("collide", d3.forceCollide().radius(d => sizeScale(d.wordCount) + 2).strength(0.8))
      .velocityDecay(0.4)
      .alphaTarget(0.2)
      .on("tick", () => {
        bubbleGroups.attr("transform", d => {
          const radius = sizeScale(d.wordCount);
          d.x = Math.max(-dimensions.width/2 + radius, Math.min(dimensions.width/2 - radius, d.x));
          d.y = Math.max(-dimensions.height/2 + radius, Math.min(dimensions.height/2 - radius, d.y));
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

    // Agency labels
    bubbleGroups.append("text")
      .attr("class", "agency-label")
      .attr("text-anchor", "middle")
      .attr("dy", "-0.5em")
      .style("fill", "#333")
      .style("font-weight", "bold")
      .style("pointer-events", "none")
      .style("stroke", "white")
      .style("stroke-width", "3px")
      .style("stroke-opacity", "0.8")
      .style("paint-order", "stroke")
      .text(d => d.agency)
      .style("font-size", d => `${Math.min(sizeScale(d.wordCount) / 4, 16)}px`);

    // Word count labels
    bubbleGroups.append("text")
      .attr("class", "count-label")
      .attr("text-anchor", "middle")
      .attr("dy", "1em")
      .style("fill", "#333")
      .style("pointer-events", "none")
      .style("stroke", "white")
      .style("stroke-width", "3px")
      .style("stroke-opacity", "0.8")
      .style("paint-order", "stroke")
      .text(d => d.wordCount.toLocaleString())
      .style("font-size", d => `${Math.min(sizeScale(d.wordCount) / 5, 14)}px`);

    // Enhanced tooltip on hover
    bubbleGroups.on("mouseover", function(event, d) {
      d3.select(this).select("circle")
        .transition()
        .duration(200)
        .style("stroke", "#333")
        .style("stroke-width", 3);
      
      // Add tooltip with full agency name
      const tooltip = container.append("g")
        .attr("class", "tooltip")
        .attr("transform", `translate(${d.x},${d.y - sizeScale(d.wordCount) - 20})`);

      tooltip.append("rect")
        .attr("fill", "white")
        .attr("rx", 4)
        .attr("ry", 4)
        .attr("opacity", 0.9);

      const text = tooltip.append("text")
        .attr("text-anchor", "middle")
        .attr("dy", "1em")
        .style("font-size", "12px")
        .text(d.fullName);

      const bbox = text.node().getBBox();
      tooltip.select("rect")
        .attr("x", bbox.x - 4)
        .attr("y", bbox.y - 4)
        .attr("width", bbox.width + 8)
        .attr("height", bbox.height + 8);
      
      simulation.current
        .force("charge", d3.forceManyBody().strength(10))
        .alpha(0.3)
        .restart();
    })
    .on("mouseout", function(event, d) {
      d3.select(this).select("circle")
        .transition()
        .duration(200)
        .style("stroke", "white")
        .style("stroke-width", 2);
      
      container.selectAll(".tooltip").remove();
      
      simulation.current
        .force("charge", d3.forceManyBody().strength(5))
        .alpha(0.2)
        .restart();
    });

    return () => {
      if (simulation.current) simulation.current.stop();
    };
  }, [data, selectedYear, dimensions]);

  return (
    <div className="w-full max-w-6xl mx-auto p-8 bg-white rounded-lg shadow-lg">
      <div className="mb-8">
        <h2 className="text-2xl font-bold mb-4">Regulatory Word Count by Agency ({selectedYear})</h2>
        {selectedYear && (
          <div className="flex items-center gap-4">
            <span className="text-sm">Year:</span>
            <input
              type="range"
              min={yearRange.min}
              max={yearRange.max}
              value={selectedYear}
              onChange={(e) => setSelectedYear(parseInt(e.target.value))}
              className="w-64"
            />
            <span className="text-sm">{selectedYear}</span>
          </div>
        )}
      </div>
      <svg
        ref={svgRef}
        width={dimensions.width}
        height={dimensions.height}
        className="bg-slate-50 rounded-lg"
      />
    </div>
  );
};

export default RegulatoryBubbles;