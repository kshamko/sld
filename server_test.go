package main

/*
func _TestCountItemsByConfig(t *testing.T) {
	app := app

	tests := []struct {
		count  int
		offset int
		want   map[ContentConfig]pagination
	}{
		{
			count:  1,
			offset: 10,
			want: map[ContentConfig]pagination{
				config1: {count: 0, offset: 6},
				config2: {count: 1, offset: 2},
				config3: {count: 0, offset: 1},
				config4: {count: 0, offset: 1},
			},
		},
		{
			count:  8,
			offset: 8,
			want: map[ContentConfig]pagination{
				config1: {count: 4, offset: 4},
				config2: {count: 2, offset: 2},
				config3: {count: 1, offset: 1},
				config4: {count: 1, offset: 1},
			},
		},
		{
			count:  1,
			offset: 0,
			want: map[ContentConfig]pagination{
				config1: {count: 1, offset: 0},
			},
		},
	}

	for _, tc := range tests {
		_, res := app.countItemsByConfig(tc.count, tc.offset)
		for cfg, items := range res {
			if tc.want[cfg].count != items.count {
				t.Errorf("cfg %+v. count expected: %d, got: %d", cfg, tc.want[cfg].count, items.count)
			}
			if tc.want[cfg].offset != items.offset {
				t.Errorf("cfg %+v. offset expected: %d, got: %d", cfg, tc.want[cfg].offset, items.offset)
			}

		}
	}
}*/
