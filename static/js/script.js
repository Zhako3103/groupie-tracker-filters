const searchInput = document.getElementById("search");
const suggestionsList = document.getElementById("suggestions");
let allArtists = [];

function updateCreationRange() {
    const minSlider = document.getElementById("creation_min");
    const maxSlider = document.getElementById("creation_max");
    const minDisplay = document.getElementById("creation_min_val");
    const maxDisplay = document.getElementById("creation_max_val");
    
    if (parseInt(minSlider.value) > parseInt(maxSlider.value)) {
        if (this === minSlider) {
            maxSlider.value = minSlider.value;
        } else {
            minSlider.value = maxSlider.value;
        }
    }
    
    minDisplay.textContent = minSlider.value;
    maxDisplay.textContent = maxSlider.value;
}

async function fetchArtists() {
    const res = await fetch("/api/artists");
    allArtists = await res.json();
}

function getSuggestions(query) {
    if (!query) return [];
    query = query.toLowerCase();
    const suggestions = [];
    const seen = new Set();
    allArtists.forEach(artist => {
        if (artist.name.toLowerCase().includes(query) && !seen.has("artist:"+artist.name)) {
            suggestions.push({
                value: artist.name,
                type: "artist/band"
            });
            seen.add("artist:"+artist.name);
        }
        // Members
        artist.members.forEach(member => {
            if (member.toLowerCase().includes(query) && !seen.has("member:"+member)) {
                suggestions.push({
                    value: member,
                    type: "member"
                });
                seen.add("member:"+member);
            }
        });
        // Locations
        if (artist.Locations) {
            artist.Locations.forEach(loc => {
                const locNorm = loc.toLowerCase().replace(/\s+/g, " ");
                if (locNorm.includes(query) && !seen.has("location:"+loc)) {
                    suggestions.push({
                        value: loc,
                        type: "location"
                    });
                    seen.add("location:"+loc);
                }
            });
        }
        // First album
        if (artist.firstAlbum && artist.firstAlbum.toLowerCase().includes(query) && !seen.has("album:"+artist.firstAlbum)) {
            suggestions.push({
                value: artist.firstAlbum,
                type: "first album"
            });
            seen.add("album:"+artist.firstAlbum);
        }
        // Creation date
        if (artist.creationDate && artist.creationDate.toString().includes(query) && !seen.has("creation:"+artist.creationDate)) {
            suggestions.push({
                value: artist.creationDate.toString(),
                type: "creation date"
            });
            seen.add("creation:"+artist.creationDate);
        }
    });
    return suggestions.slice(0, 8); // limit
}

function renderSuggestions(suggestions) {
    suggestionsList.innerHTML = "";
    suggestions.forEach(s => {
        const li = document.createElement("li");
        li.className = "suggestion-item";
        li.setAttribute("role", "option");
        li.innerHTML = `<span class="suggestion-value">${s.value}</span> <span class="suggestion-type">- ${s.type}</span>`;
        li.addEventListener("mousedown", e => {
            e.preventDefault();
            searchInput.value = s.value;
            suggestionsList.innerHTML = "";
            // Optionally trigger search
            searchInput.dispatchEvent(new Event("input"));
        });
        suggestionsList.appendChild(li);
    });
    suggestionsList.style.display = suggestions.length ? "block" : "none";
}

searchInput.addEventListener("input", () => {
    const query = searchInput.value.toLowerCase();
    const filtered = allArtists.filter(a =>
        a.name.toLowerCase().includes(query) ||
        a.members.some(m => m.toLowerCase().includes(query)) ||
        (a.Locations && a.Locations.some(l => l.toLowerCase().includes(query))) ||
        (a.firstAlbum && a.firstAlbum.toLowerCase().includes(query)) ||
        (a.creationDate && a.creationDate.toString().includes(query))
    );
    startTetris(filtered.length ? filtered : allArtists);
    renderSuggestions(getSuggestions(searchInput.value));
});

searchInput.addEventListener("blur", () => {
    setTimeout(() => suggestionsList.innerHTML = "", 100);
});

fetchArtists();
